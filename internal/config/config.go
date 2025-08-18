package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type Provider struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
}

type Config struct {
	ThemeName         string     `json:"theme_name"`
	SelectedInterface string     `json:"selected_interface"`
	Providers         []Provider `json:"providers"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)

	if err != nil {
		if os.IsNotExist(err) {
			defaultConfig := &Config{
				ThemeName: "Default",
				Providers: GetDefaultProviders(),
			}

			if err := SaveConfig(path, defaultConfig); err != nil {
				return nil, fmt.Errorf("could not create default config file: %w", err)
			}

			return defaultConfig, nil
		}

		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config

	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file, maybe it's corrupted: %w", err)
	}

	if cfg.ThemeName == "" {
		cfg.ThemeName = "Default"
	}

	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	sort.Slice(cfg.Providers, func(i, j int) bool {
		return cfg.Providers[i].Name < cfg.Providers[j].Name
	})

	data, err := json.MarshalIndent(cfg, "", "\t")

	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error saving config file: %w", err)
	}

	return nil
}

func GetDefaultProviders() []Provider {
	return []Provider{
		{Name: "Shecan", Addresses: []string{"178.22.122.100", "185.51.200.2"}},
		{Name: "Electro", Addresses: []string{"78.157.42.100", "78.157.42.101"}},
		{Name: "Radar Game", Addresses: []string{"10.202.10.11", "10.202.10.12"}},
		{Name: "Control D (Gaming)", Addresses: []string{"76.76.2.22", "76.76.10.22"}},
		{Name: "Cloudflare", Addresses: []string{"1.1.1.1", "1.0.0.1"}},
		{Name: "Google", Addresses: []string{"8.8.8.8", "8.8.4.4"}},
		{Name: "Quad9", Addresses: []string{"9.9.9.9", "149.112.112.112"}},
		{Name: "OpenDNS", Addresses: []string{"208.67.222.222", "208.67.220.220"}},
		{Name: "Neustar UltraDNS", Addresses: []string{"156.154.70.1", "156.154.71.1"}},
		{Name: "DNS.WATCH", Addresses: []string{"84.200.69.80", "84.200.70.40"}},
		{Name: "Comodo Secure", Addresses: []string{"8.26.56.26", "8.20.247.20"}},
		{Name: "Level3", Addresses: []string{"209.244.0.3", "209.244.0.4"}},
		{Name: "Yandex", Addresses: []string{"77.88.8.8", "77.88.8.1"}},
		{Name: "AdGuard DNS", Addresses: []string{"94.140.14.14", "94.140.15.15"}},
		{Name: "AdGuard DNS (Family)", Addresses: []string{"94.140.14.15", "94.140.15.16"}},
		{Name: "Mullvad DNS (Ad Block)", Addresses: []string{"194.242.2.3", "193.19.108.3"}},
		{Name: "AlternateDNS", Addresses: []string{"76.76.19.19", "76.223.122.150"}},
		{Name: "CleanBrowse (Security)", Addresses: []string{"185.228.168.9", "185.228.169.9"}},
		{Name: "Cloudflare (Malware Block)", Addresses: []string{"1.1.1.2", "1.0.0.2"}},
		{Name: "Quad9 (Secure)", Addresses: []string{"9.9.9.10", "149.112.112.10"}},
		{Name: "OpenDNS FamilyShield", Addresses: []string{"208.67.222.123", "208.67.220.123"}},
	}
}
