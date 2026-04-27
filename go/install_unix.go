//go:build !windows
// +build !windows

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type nativeManifest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Path           string   `json:"path"`
	Type           string   `json:"type"`
	AllowedOrigins []string `json:"allowed_origins"`
}

func defaultInstallDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "free-quick-share"), nil
}

func installedBinaryName() string {
	return "free-quick-share-host"
}

func registerNativeHost(hostNameValue, binaryPath, extensionID, browser string) ([]string, error) {
	manifestDir, err := nativeManifestDir(browser)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(manifestDir, hostNameValue+".json")

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
	return []string{manifestPath}, nil
}

func unregisterNativeHost(hostNameValue, browser string) error {
	manifestDirs, err := nativeManifestDirs(browser, true)
	if err != nil {
		return err
	}
	for _, dir := range manifestDirs {
		manifestPath := filepath.Join(dir, hostNameValue+".json")
		if err := os.Remove(manifestPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func nativeManifestDir(browser string) (string, error) {
	dirs, err := nativeManifestDirs(browser, false)
	if err != nil {
		return "", err
	}
	return dirs[0], nil
}

func nativeManifestDirs(browser string, allowAll bool) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	var chromeDir string
	var chromiumDir string
	if runtime.GOOS == "darwin" {
		chromeDir = filepath.Join(home, "Library", "Application Support", "Google", "Chrome", "NativeMessagingHosts")
		chromiumDir = filepath.Join(home, "Library", "Application Support", "Chromium", "NativeMessagingHosts")
	} else {
		chromeDir = filepath.Join(home, ".config", "google-chrome", "NativeMessagingHosts")
		chromiumDir = filepath.Join(home, ".config", "chromium", "NativeMessagingHosts")
	}

	switch browser {
	case "chrome":
		return []string{chromeDir}, nil
	case "chromium":
		return []string{chromiumDir}, nil
	case "all":
		if allowAll {
			return []string{chromeDir, chromiumDir}, nil
		}
	}
	return nil, fmt.Errorf("unsupported browser: %s", browser)
}

func killResidualBinaryProcesses(binaryPath string) {
	if _, err := exec.LookPath("pkill"); err != nil {
		return
	}
	_ = exec.Command("pkill", "-f", binaryPath).Run()
}

func removeInstallDir(installDir string) error {
	return os.RemoveAll(installDir)
}
