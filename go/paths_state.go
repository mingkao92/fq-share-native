package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func resolvePaths() (appPaths, error) {
	selfPath, err := os.Executable()
	if err != nil {
		return appPaths{}, err
	}
	selfPath, err = filepath.Abs(selfPath)
	if err != nil {
		return appPaths{}, err
	}
	baseDir := filepath.Dir(selfPath)
	runtimeDir := filepath.Join(baseDir, "runtime")
	dataDir := filepath.Join(baseDir, "data")

	return appPaths{
		selfPath:    selfPath,
		baseDir:     baseDir,
		runtimeDir:  runtimeDir,
		dataDir:     dataDir,
		statePath:   filepath.Join(runtimeDir, "state.json"),
		httpLogPath: filepath.Join(runtimeDir, "http_server.log"),
		hostLogPath: filepath.Join(runtimeDir, "host_error.log"),
	}, nil
}

func pathsForInstallDir(baseDir, binaryPath string) appPaths {
	runtimeDir := filepath.Join(baseDir, "runtime")
	dataDir := filepath.Join(baseDir, "data")
	return appPaths{
		selfPath:    binaryPath,
		baseDir:     baseDir,
		runtimeDir:  runtimeDir,
		dataDir:     dataDir,
		statePath:   filepath.Join(runtimeDir, "state.json"),
		httpLogPath: filepath.Join(runtimeDir, "http_server.log"),
		hostLogPath: filepath.Join(runtimeDir, "host_error.log"),
	}
}

func ensureDirs(paths appPaths) error {
	dirs := []string{
		paths.runtimeDir,
		paths.dataDir,
		filepath.Join(paths.dataDir, "sessions"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func loadState(paths appPaths) *serverState {
	data, err := os.ReadFile(paths.statePath)
	if err != nil {
		return nil
	}
	var state serverState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil
	}
	if !isPIDRunning(state.PID) {
		if state.Port > 0 && waitHealth(state.Port, time.Second) {
			return &state
		}
		_ = clearState(paths)
		return nil
	}
	return &state
}

func saveState(paths appPaths, state *serverState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(paths.statePath, data, 0o644)
}

func clearState(paths appPaths) error {
	if err := os.Remove(paths.statePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func logHostError(path, message string) {
	line := fmt.Sprintf("[%d] %s\n", time.Now().Unix(), message)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line)
}
