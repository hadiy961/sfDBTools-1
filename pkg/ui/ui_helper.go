// File : pkg/ui/ui_helper.go
// Deskripsi : Fungsi utilitas untuk operasi UI di terminal
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sfDBTools/internal/appconfig"
	app_log "sfDBTools/internal/applog"
)

// ClearScreen clears the terminal screen using platform-specific commands
func ClearScreen() error {
	lg := app_log.NewLogger()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		// Linux, macOS, and other Unix-like systems
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		lg.Warnf("Failed to clear screen using system command, falling back to ANSI escape sequences - os=%s: %v", runtime.GOOS, err)
		return ClearScreenANSI()
	}

	// Only log if there are issues, not for successful operations
	return nil
}

// ClearScreenANSI clears the terminal screen using ANSI escape sequences
func ClearScreenANSI() error {
	lg := app_log.NewLogger()

	// ANSI escape sequence to clear screen and move cursor to top-left
	_, err := fmt.Print("\033[2J\033[H")
	if err != nil {
		lg.Errorf("Failed to clear screen using ANSI escape sequences: %v", err)
		return err
	}

	// Only log if there are issues, not for successful operations
	return nil
}

// ClearWithMessage clears screen and displays a message
func ClearWithMessage(message string) error {
	if err := ClearScreen(); err != nil {
		return err
	}
	if message != "" {
		fmt.Println(message)
	}
	return nil
}

// ClearAndShowHeader clears screen and shows a formatted header
func ClearAndShowHeader(title string) error {
	if err := ClearScreen(); err != nil {
		return err
	}
	PrintHeader(title)
	return nil
}

func Headers(title string) {
	cfg, err := appconfig.LoadConfigFromEnv()
	if err != nil {
		PrintError(fmt.Sprintf("Error loading config: %v", err))
		return
	}

	ClearAndShowHeader(cfg.General.AppName + " v" + cfg.General.Version + " - " + title)
}
