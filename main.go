package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"swift-dns/internal/config"
	"swift-dns/internal/dns"
	"swift-dns/internal/platform"
	"swift-dns/internal/platform/windows"
	"swift-dns/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var p platform.Platform

	switch runtime.GOOS {
	case "windows":
		p = windows.New()
	default:
		log.Fatalf("Unsupported operating system: %s", runtime.GOOS)
	}

	if err := p.CheckAdmin(); err != nil {
		fmt.Println(err)
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
		os.Exit(1)
	}

	devMode := flag.Bool("dev", false, "Enable development mode to create a debug.log file")
	flag.Bool("admin", false, "Internal use for admin check")

	flag.Parse()

	if *devMode {
		logFile, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)

		if err != nil {
			log.Fatalf("error opening log file: %v", err)
		}

		defer logFile.Close()

		log.SetOutput(logFile)
		log.Println("Dev mode enabled, logging to debug.log")
	} else {
		log.SetOutput(io.Discard)
	}

	configPath, err := p.GetConfigPath()

	if err != nil {
		log.Fatalf("Could not determine config path: %v", err)
	}

	cfg, err := config.LoadConfig(configPath)

	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	var dnsManager dns.Manager

	if cfg.SelectedInterface != "" {
		dnsManager, err = p.NewDNSManager(cfg.SelectedInterface)

		if err != nil {
			cfg.SelectedInterface = ""
			config.SaveConfig(configPath, cfg)
		}
	}

	initialModel := tui.InitialModel(p, configPath, dnsManager, cfg)
	program := tea.NewProgram(&initialModel, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
