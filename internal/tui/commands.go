package tui

import (
	"sort"
	"swift-dns/internal/config"
	"swift-dns/internal/ping"
	"swift-dns/internal/platform"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func saveConfigCmd(m model, successMsg string) tea.Cmd {
	return func() tea.Msg {
		cfg := &config.Config{
			ThemeName:         m.currentThemeName,
			SelectedInterface: m.selectedInterface,
			Providers:         m.providers,
		}

		if err := config.SaveConfig(m.configPath, cfg); err != nil {
			return errorMsg{err}
		}

		return listUpdatedMsg(successMsg)
	}
}

func checkStatusCmd(m model) tea.Cmd {
	return func() tea.Msg {
		activeDNS, err := m.dnsManager.GetActive(m.providers)

		if err != nil {
			return errorMsg{err}
		}

		return statusMsg(activeDNS)
	}
}

func setDNSCmd(m model, provider config.Provider) tea.Cmd {
	return func() tea.Msg {
		if err := m.dnsManager.Set(provider); err != nil {
			return errorMsg{err}
		}

		return dnsSetMsg{provider: provider}
	}
}

func clearDNSCmd(m model) tea.Cmd {
	return func() tea.Msg {
		if err := m.dnsManager.Clear(); err != nil {
			return errorMsg{err}
		}

		return dnsClearedMsg("DNS settings successfully cleared (set to DHCP).")
	}
}

func pingAllCmd(providers []config.Provider) tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup

		resultsChan := make(chan ping.JobResult, len(providers))

		for _, provider := range providers {
			wg.Add(1)

			go func(p config.Provider) {
				defer wg.Done()

				if len(p.Addresses) > 0 {
					res, _ := ping.Ping(p.Addresses[0])
					resultsChan <- ping.JobResult{Provider: p, Result: res}
				}
			}(provider)
		}

		wg.Wait()
		close(resultsChan)

		var results []ping.JobResult

		for result := range resultsChan {
			results = append(results, result)
		}

		sort.Slice(results, func(i, j int) bool {
			if results[i].Result.TimedOut {
				return false
			}

			if results[j].Result.TimedOut {
				return true
			}

			return results[i].Result.AvgRtt < results[j].Result.AvgRtt
		})

		return pingResultsMsg(results)
	}
}

func loadInterfacesCmd(p platform.Platform) tea.Cmd {
	return func() tea.Msg {
		interfaces, err := p.GetNetworkInterfaces()

		if err != nil {
			return errorMsg{err}
		}

		return interfacesLoadedMsg(interfaces)
	}
}

func clearMessageAfterCmd(id int, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(duration)
		return clearMessageMsg{id: id}
	}
}
