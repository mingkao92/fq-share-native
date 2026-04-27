package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func runNativeHost() error {
	paths, err := resolvePaths()
	if err != nil {
		return err
	}
	if err := ensureDirs(paths); err != nil {
		return err
	}

	for {
		msg, err := readNativeMessage(os.Stdin)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			logHostError(paths.hostLogPath, fmt.Sprintf("read message failed: %v", err))
			continue
		}

		resp := handleMessage(paths, msg)
		if err := writeNativeMessage(os.Stdout, resp); err != nil {
			logHostError(paths.hostLogPath, fmt.Sprintf("write message failed: %v", err))
			return err
		}
	}
}

func readNativeMessage(r io.Reader) (map[string]any, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(r, header); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, io.EOF
		}
		return nil, err
	}

	length := binary.LittleEndian.Uint32(header)
	if length == 0 {
		return map[string]any{}, nil
	}
	if length > 10*1024*1024 {
		return nil, fmt.Errorf("message too large: %d", length)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, io.EOF
		}
		return nil, err
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return map[string]any{"action": "invalid"}, nil
	}
	return payload, nil
}

func writeNativeMessage(w io.Writer, payload map[string]any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	header := make([]byte, 4)
	binary.LittleEndian.PutUint32(header, uint32(len(encoded)))
	if _, err := w.Write(header); err != nil {
		return err
	}
	if _, err := w.Write(encoded); err != nil {
		return err
	}
	return nil
}

func handleMessage(paths appPaths, message map[string]any) map[string]any {
	requestID := message["id"]
	action, _ := message["action"].(string)

	var response map[string]any
	switch action {
	case "start_server":
		response = withNativeLock(paths, startServer)
	case "stop_server":
		response = withNativeLock(paths, stopServer)
	case "get_state":
		response = withNativeLock(paths, func(p appPaths) map[string]any {
			return buildStatus(p, nil)
		})
	default:
		response = map[string]any{"ok": false, "error": fmt.Sprintf("Unknown action: %s", action)}
	}
	response["id"] = requestID
	return response
}

func withNativeLock(paths appPaths, fn func(appPaths) map[string]any) map[string]any {
	lockPath := paths.statePath + ".lock"
	deadline := time.Now().Add(3 * time.Second)
	for {
		lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = lockFile.WriteString(strconv.Itoa(os.Getpid()))
			_ = lockFile.Close()
			defer os.Remove(lockPath)
			return fn(paths)
		}
		if !os.IsExist(err) {
			return map[string]any{"ok": false, "error": fmt.Sprintf("acquire lock failed: %v", err)}
		}

		// Guard against stale lock files from abrupt process exits.
		if info, statErr := os.Stat(lockPath); statErr == nil && time.Since(info.ModTime()) > 12*time.Second {
			_ = os.Remove(lockPath)
			continue
		}
		if time.Now().After(deadline) {
			return map[string]any{"ok": false, "error": "native operation busy, please retry"}
		}
		time.Sleep(80 * time.Millisecond)
	}
}

func startServer(paths appPaths) map[string]any {
	state := loadState(paths)
	if state != nil {
		return buildStatus(paths, state)
	}

	port, err := findFreePort(defaultPort, maxPort)
	if err != nil {
		return map[string]any{"ok": false, "error": err.Error()}
	}
	token, err := generateToken(18)
	if err != nil {
		return map[string]any{"ok": false, "error": err.Error()}
	}
	lanIP := detectLANIP()

	logFile, err := os.OpenFile(paths.httpLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return map[string]any{"ok": false, "error": fmt.Sprintf("open log failed: %v", err)}
	}
	defer logFile.Close()

	cmd := exec.Command(
		paths.selfPath,
		"serve-http",
		"--port", strconv.Itoa(port),
		"--token", token,
		"--data-dir", paths.dataDir,
	)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	cmd.Dir = paths.baseDir
	configureChildProcess(cmd)

	if err := cmd.Start(); err != nil {
		return map[string]any{"ok": false, "error": fmt.Sprintf("spawn failed: %v", err)}
	}

	if !waitHealth(port, healthWaitTime) {
		terminateForce(cmd.Process.Pid)
		return map[string]any{"ok": false, "error": "Failed to start http server"}
	}

	newState := &serverState{
		PID:       cmd.Process.Pid,
		Port:      port,
		Token:     token,
		LANIP:     lanIP,
		StartedAt: time.Now().Unix(),
	}
	if err := saveState(paths, newState); err != nil {
		terminateForce(cmd.Process.Pid)
		return map[string]any{"ok": false, "error": fmt.Sprintf("save state failed: %v", err)}
	}

	return buildStatus(paths, newState)
}

func stopServer(paths appPaths) map[string]any {
	state := loadState(paths)
	if state == nil {
		return buildStatus(paths, nil)
	}

	terminateGraceful(state.PID)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !isPIDRunning(state.PID) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if isPIDRunning(state.PID) {
		terminateForce(state.PID)
	}

	_ = clearState(paths)
	return buildStatus(paths, nil)
}

func stopManagedServer(paths appPaths) {
	state := loadState(paths)
	if state == nil {
		killResidualBinaryProcesses(paths.selfPath)
		return
	}

	terminateGraceful(state.PID)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !isPIDRunning(state.PID) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if isPIDRunning(state.PID) {
		terminateForce(state.PID)
	}
	killResidualBinaryProcesses(paths.selfPath)
	_ = clearState(paths)
}

func buildStatus(paths appPaths, state *serverState) map[string]any {
	if state == nil {
		state = loadState(paths)
	}
	if state == nil {
		return map[string]any{
			"ok":        true,
			"running":   false,
			"host_name": hostName,
		}
	}

	lanURL := fmt.Sprintf("http://%s:%d", state.LANIP, state.Port)
	return map[string]any{
		"ok":         true,
		"running":    true,
		"host_name":  hostName,
		"pid":        state.PID,
		"port":       state.Port,
		"token":      state.Token,
		"lan_ip":     state.LANIP,
		"local_url":  fmt.Sprintf("http://127.0.0.1:%d", state.Port),
		"lan_url":    lanURL,
		"phone_url":  fmt.Sprintf("%s/?token=%s", lanURL, url.QueryEscape(state.Token)),
		"started_at": state.StartedAt,
	}
}
