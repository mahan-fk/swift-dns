package windows

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"swift-dns/internal/dns"
)

type winPlatform struct{}

func New() *winPlatform {
	return &winPlatform{}
}

func (w *winPlatform) GetConfigPath() (string, error) {
	var appDataDir = os.Getenv("APPDATA")

	if appDataDir == "" {
		return "", fmt.Errorf("APPDATA environment variable is not set")
	}

	configDir := filepath.Join(appDataDir, "swift-dns")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("error creating config directory: %w", err)
	}

	return filepath.Join(configDir, "dns_config.json"), nil
}

func (w *winPlatform) ClearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

func (p *winPlatform) Pause() {
	fmt.Println("\nPress Enter to continue...")
	_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func (w *winPlatform) CheckAdmin() error {
	if len(os.Args) > 1 && os.Args[1] == "--admin" {
		return nil
	}

	cmd := exec.Command("net", "session")

	if cmd.Run() == nil {
		return nil
	}

	cmd = exec.Command("powershell", "Start-Process", "-FilePath", os.Args[0], "-ArgumentList", "'--admin'", "-Verb", "RunAs")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to elevate to administrator: %w\nPlease run this program as an administrator", err)
	}

	os.Exit(0)

	return nil
}

func (w *winPlatform) GetNetworkInterfaces() ([]string, error) {
	cmd := exec.Command("powershell", "-Command", "Get-NetAdapter | Where-Object {$_.Status -eq 'Up'} | Select-Object -ExpandProperty Name")

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %s, %w", stderr.String(), err)
	}

	interfaces := strings.Split(strings.TrimSpace(out.String()), "\r\n")

	if len(interfaces) == 0 || (len(interfaces) == 1 && interfaces[0] == "") {
		return nil, fmt.Errorf("no active network interfaces found")
	}

	return interfaces, nil
}

func (w *winPlatform) NewDNSManager(interfaceName string) (dns.Manager, error) {
	return &winDNSManager{interfaceName: interfaceName}, nil
}
