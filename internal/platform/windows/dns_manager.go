package windows

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"swift-dns/internal/config"
)

type winDNSManager struct {
	interfaceName string
}

func (m *winDNSManager) GetActive(providers []config.Provider) (string, error) {
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("(Get-DnsClientServerAddress -InterfaceAlias \"%s\" -AddressFamily IPv4).ServerAddresses", m.interfaceName))
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("could not get active DNS: %w. Make sure the network interface '%s' exists", err, m.interfaceName)
	}

	activeDNSList := strings.Fields(string(output))

	if len(activeDNSList) == 1 {
		return "DHCP [Automatic]", nil
	}

	sort.Strings(activeDNSList)

	for _, provider := range providers {
		providerAddressesCopy := make([]string, len(provider.Addresses))
		copy(providerAddressesCopy, provider.Addresses)
		sort.Strings(providerAddressesCopy)

		if strings.Join(activeDNSList, ",") == strings.Join(providerAddressesCopy, ",") {
			return fmt.Sprintf("%s [%s]", provider.Name, strings.Join(provider.Addresses, " | ")), nil
		}
	}

	return fmt.Sprintf("Custom [%s]", strings.Join(activeDNSList, " | ")), nil
}

func (m *winDNSManager) Set(provider config.Provider) error {
	if len(provider.Addresses) < 1 {
		return fmt.Errorf("no DNS addresses provided for %s", provider.Name)
	}

	primaryCmd := exec.Command("netsh", "interface", "ip", "set", "dns", fmt.Sprintf("name=\"%s\"", m.interfaceName), "static", provider.Addresses[0])

	if err := primaryCmd.Run(); err != nil {
		return fmt.Errorf("error setting primary DNS: %w", err)
	}

	if len(provider.Addresses) > 1 {
		secondaryDNS := exec.Command("netsh", "interface", "ip", "add", "dns", fmt.Sprintf("name=\"%s\"", m.interfaceName), provider.Addresses[1], "index=2")

		if err := secondaryDNS.Run(); err != nil {
			_ = m.Clear()
			return fmt.Errorf("error setting secondary DNS: %w", err)
		}
	}

	return m.Flush()
}

func (m *winDNSManager) Clear() error {
	cmd := exec.Command("netsh", "interface", "ip", "set", "dns", fmt.Sprintf("name=\"%s\"", m.interfaceName), "dhcp")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error clearing DNS settings: %w", err)
	}

	return m.Flush()
}

func (m *winDNSManager) Flush() error {
	flushCmd := exec.Command("ipconfig", "/flushdns")

	if err := flushCmd.Run(); err != nil {
		return fmt.Errorf("could not flush the DNS cache: %w", err)
	}

	return nil
}
