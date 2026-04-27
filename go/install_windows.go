//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type nativeManifest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Path           string   `json:"path"`
	Type           string   `json:"type"`
	AllowedOrigins []string `json:"allowed_origins"`
}

func defaultInstallDir() (string, error) {
	localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
	if localAppData == "" {
		return "", fmt.Errorf("LOCALAPPDATA is not set")
	}
	return filepath.Join(localAppData, "FreeQuickShare"), nil
}

func installedBinaryName() string {
	return "free-quick-share-host.exe"
}

func registerNativeHost(hostNameValue, binaryPath, extensionID, browser string) ([]string, error) {
	manifestPath := filepath.Join(filepath.Dir(binaryPath), hostNameValue+".json")
	manifest := nativeManifest{
		Name:        hostNameValue,
		Description: "Free Quick Share Native Host",
		Path:        binaryPath,
		Type:        "stdio",
		// Keep previously registered extension ids so unpacked + CRX installs can coexist.
		AllowedOrigins: buildAllowedOrigins(manifestPath, extensionID),
	}
	payload, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(manifestPath, payload, 0o644); err != nil {
		return nil, err
	}

	regPath := windowsNativeHostRegPath(hostNameValue, browser)
	cmd := exec.Command("reg", "add", regPath, "/ve", "/t", "REG_SZ", "/d", manifestPath, "/f")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to register host: %v (%s)", err, strings.TrimSpace(string(output)))
	}
	return []string{manifestPath, regPath}, nil
}

func unregisterNativeHost(hostNameValue, browser string) error {
	installDir, err := defaultInstallDir()
	if err != nil {
		return err
	}
	manifestPath := filepath.Join(installDir, hostNameValue+".json")
	if err := os.Remove(manifestPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	regPaths := windowsNativeHostRegPaths(hostNameValue, browser)
	for _, regPath := range regPaths {
		cmd := exec.Command("reg", "delete", regPath, "/f")
		output, err := cmd.CombinedOutput()
		if err == nil {
			continue
		}
		lower := strings.ToLower(string(output))
		if strings.Contains(lower, "unable to find") || strings.Contains(lower, "cannot find") {
			continue
		}
		return fmt.Errorf("failed to remove registry key %s: %v (%s)", regPath, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func windowsNativeHostRegPath(hostNameValue, browser string) string {
	if browser == "chromium" {
		return fmt.Sprintf(`HKCU\Software\Chromium\NativeMessagingHosts\%s`, hostNameValue)
	}
	return fmt.Sprintf(`HKCU\Software\Google\Chrome\NativeMessagingHosts\%s`, hostNameValue)
}

func windowsNativeHostRegPaths(hostNameValue, browser string) []string {
	if browser == "chrome" {
		return []string{windowsNativeHostRegPath(hostNameValue, "chrome")}
	}
	if browser == "chromium" {
		return []string{windowsNativeHostRegPath(hostNameValue, "chromium")}
	}
	return []string{
		windowsNativeHostRegPath(hostNameValue, "chrome"),
		windowsNativeHostRegPath(hostNameValue, "chromium"),
	}
}

func killResidualBinaryProcesses(_ string) {
	// State-based PID cleanup is the primary mechanism on Windows.
}

func removeInstallDir(installDir string) error {
	for i := 0; i < 6; i++ {
		err := os.RemoveAll(installDir)
		if err == nil || os.IsNotExist(err) {
			return nil
		}
		time.Sleep(400 * time.Millisecond)
	}
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return nil
	}
	return fmt.Errorf("failed to remove install dir: %s", installDir)
}
