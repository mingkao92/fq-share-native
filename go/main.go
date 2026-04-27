package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "serve-http":
			if err := runHTTPMode(os.Args[2:]); err != nil {
				log.Fatalf("serve-http failed: %v", err)
			}
			return
		case "install":
			if err := runInstallMode(os.Args[2:]); err != nil {
				log.Fatalf("install failed: %v", err)
			}
			return
		case "uninstall":
			if err := runUninstallMode(os.Args[2:]); err != nil {
				log.Fatalf("uninstall failed: %v", err)
			}
			return
		}
	}

	if err := runNativeHost(); err != nil {
		log.Printf("native host exited with error: %v", err)
		os.Exit(1)
	}
}
