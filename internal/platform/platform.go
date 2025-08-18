package platform

import "swift-dns/internal/dns"

type Platform interface {
	GetConfigPath() (string, error)
	ClearScreen()
	Pause()
	CheckAdmin() error
	GetNetworkInterfaces() ([]string, error)
	NewDNSManager(interfaceName string) (dns.Manager, error)
}
