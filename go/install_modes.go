package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func runInstallMode(args []string) error {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	extensionID := fs.String("extension-id", "", "")
	browser := fs.String("browser", "chrome", "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*extensionID) == "" {
		return errors.New("missing --extension-id")
	}
	if err := validateBrowser(*browser, false); err != nil {
		return err
	}

	installDir, err := defaultInstallDir()
	if err != nil {
		return err
	}
	targetBinary := filepath.Join(installDir, installedBinaryName())
	paths := pathsForInstallDir(installDir, targetBinary)

	if err := ensureDirs(paths); err != nil {
		return err
	}
	stopManagedServer(paths)

	sourceBinary, err := os.Executable()
	if err != nil {
		return err
	}
	if err := copyBinary(sourceBinary, targetBinary); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(targetBinary, 0o755); err != nil {
			return err
		}
	}

	// Best-effort cleanup for previously used host names.
	for _, legacyHost := range legacyHostNames {
		if legacyHost == hostName {
			continue
		}
		_ = unregisterNativeHost(legacyHost, *browser)
	}

	installedItems, err := registerNativeHost(hostName, targetBinary, *extensionID, *browser)
	if err != nil {
		return err
	}

	fmt.Printf("Installed native host: %s\n", hostName)
	fmt.Printf("Binary: %s\n", targetBinary)
	for _, item := range installedItems {
		fmt.Printf("Registered: %s\n", item)
	}
	return nil
}

func runUninstallMode(args []string) error {
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	browser := fs.String("browser", "all", "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := validateBrowser(*browser, true); err != nil {
		return err
	}

	installDir, err := defaultInstallDir()
	if err != nil {
		return err
	}
	targetBinary := filepath.Join(installDir, installedBinaryName())
	paths := pathsForInstallDir(installDir, targetBinary)

	stopManagedServer(paths)

	if err := unregisterNativeHost(hostName, *browser); err != nil {
		return err
	}
	for _, legacyHost := range legacyHostNames {
		if legacyHost == hostName {
			continue
		}
		_ = unregisterNativeHost(legacyHost, *browser)
	}

	currentExe, _ := os.Executable()
	if runtime.GOOS == "windows" && filepath.Clean(currentExe) == filepath.Clean(targetBinary) {
		return errors.New("cannot remove install directory while running installed binary; run uninstall from a temporary downloaded binary")
	}
	if err := removeInstallDir(installDir); err != nil {
		return err
	}

	fmt.Printf("Uninstalled native host: %s\n", hostName)
	fmt.Printf("Removed install dir: %s\n", installDir)
	return nil
}

func validateBrowser(browser string, allowAll bool) error {
	switch browser {
	case "chrome", "chromium":
		return nil
	case "all":
		if allowAll {
			return nil
		}
	}
	if allowAll {
		return fmt.Errorf("invalid browser: %s (expected chrome|chromium|all)", browser)
	}
	return fmt.Errorf("invalid browser: %s (expected chrome|chromium)", browser)
}

func copyBinary(srcPath, dstPath string) error {
	srcAbs, err := filepath.Abs(srcPath)
	if err != nil {
		return err
	}
	dstAbs, err := filepath.Abs(dstPath)
	if err != nil {
		return err
	}
	if srcAbs == dstAbs {
		return nil
	}

	srcFile, err := os.Open(srcAbs)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dstAbs), 0o755); err != nil {
		return err
	}
	dstFile, err := os.OpenFile(dstAbs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return dstFile.Sync()
}
