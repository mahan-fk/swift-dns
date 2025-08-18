package tui

import (
	"errors"
	"fmt"
	"strings"
	"swift-dns/internal/config"
	"swift-dns/internal/ping"
	"swift-dns/internal/utils"

	"github.com/charmbracelet/bubbles/textinput"
)

func renderMainMenuItem(item any, isSelected bool, theme *Theme) string {
	label, ok := item.(string)

	if !ok {
		return "Error: Invalid item type for Main Menu"
	}

	line := fmt.Sprintf("  %s", label)

	if isSelected {
		return theme.Selected.Render(line)
	}

	return line
}

func renderDNSItem(item any, isSelected bool, theme *Theme) string {
	provider, ok := item.(config.Provider)

	if !ok {
		return "Error: Invalid item type for DNS list"
	}

	line := fmt.Sprintf("  %-40s [%s]", provider.Name, strings.Join(provider.Addresses, " | "))

	if isSelected {
		return theme.Selected.Render(line)
	}

	return line
}

func renderInterfaceItem(item any, isSelected bool, theme *Theme) string {
	iface, ok := item.(string)

	if !ok {
		return "Error: Invalid item type for Interface selection"
	}

	line := fmt.Sprintf("  %s", iface)

	if isSelected {
		return theme.Selected.Render(line)
	}

	return line
}

func renderPingResultItem(item any, isSelected bool, theme *Theme) string {
	result, ok := item.(ping.JobResult)

	if !ok {
		return "Error: Invalid item type for ping result"
	}

	var pingStr string

	if result.Result.TimedOut {
		pingStr = theme.PingTimeout.Render("Timed Out")
	} else {
		pingStr = theme.PingResult.Render(fmt.Sprintf("%dms", result.Result.AvgRtt.Milliseconds()))
	}

	return fmt.Sprintf("  %-40s Ping: %s", result.Provider.Name, pingStr)
}

func renderThemeItem(item any, isSelected bool, theme *Theme) string {
	themeName, ok := item.(string)

	if !ok {
		return "Error: Invalid item type for Theme selection"
	}

	line := fmt.Sprintf("  %s", themeName)

	if isSelected {
		return theme.Selected.Render(line)
	}

	return line
}

func viewHeader(m model) string {
	var messageContent string

	if m.message != "" {
		messageContent = m.message
	} else {
		messageContent = m.currentTheme.Help.Render("---")
	}

	var statusLine string

	if m.selectedInterface != "" {
		statusLine = fmt.Sprintf(
			"%s %s on %s",
			m.currentTheme.StatusHeader.Render("Current Status:"),
			m.currentTheme.Status.Render(m.status),
			m.currentTheme.Title.Render(m.selectedInterface),
		)
	} else {
		statusLine = fmt.Sprintf(
			"%s %s",
			m.currentTheme.StatusHeader.Render("Current Status:"),
			m.currentTheme.Help.Render(m.status),
		)
	}

	return fmt.Sprintf(
		"%s\n%s\n%s %s\n%s %s\n",
		m.currentTheme.Title.Render("📡 Swift DNS"),
		statusLine,
		m.currentTheme.StatusHeader.Render("Theme:"),
		m.currentTheme.Help.Render(m.currentThemeName),
		m.currentTheme.StatusHeader.Render("Message:"),
		messageContent,
	)
}

func viewLoading(m model) string {
	return fmt.Sprintf("%s %s\n", m.spinner.View(), m.loadingMessage)
}

func viewConfirm(m model, action, target string) string {
	var b strings.Builder

	b.WriteString(m.currentTheme.Warning.Render(fmt.Sprintf("⚠️ Are you sure you want to %s %s?", action, target)) + "\n\n")
	b.WriteString("  (y) Yes, proceed\n")
	b.WriteString("  (n) No, go back\n")

	return b.String()
}

func viewForm(m model) string {
	var b strings.Builder

	title := "➕ Add a new DNS Provider"

	if m.state == updateDNSFormView {
		title = fmt.Sprintf("✏️ Update '%s'", m.providers[m.providerToActOn].Name)
	}

	b.WriteString(title + "\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View() + "\n")
	}

	return b.String()
}

func viewFooter(m model) string {
	var parts []string

	theme := m.currentTheme

	switch m.state {
	case mainMenuView:
		parts = []string{"↑/↓", "navigate", "←/→", "pages", "enter", "select", "q", "quit"}
	case setDNSListView, updateDNSListView, deleteDNSListView, themeSelectionView:
		parts = []string{"↑/↓", "navigate", "←/→", "pages", "enter", "select", "esc/q", "back"}
	case interfaceSelectionView:
		if m.dnsManager == nil {
			parts = []string{"↑/↓", "navigate", "←/→", "pages", "enter", "select", "q", "quit"}
		} else {
			parts = []string{"↑/↓", "navigate", "←/→", "pages", "enter", "select", "esc/q", "back"}
		}
	case addDNSView, updateDNSFormView:
		parts = []string{"tab", "next field", "ctrl+s/enter", "save", "esc/q", "back"}
	case pingingView:
		parts = []string{"←/→", "pages", "esc/q", "back"}
	case confirmDeleteView, confirmResetView:
		parts = []string{"y", "yes", "n", "no", "esc/q", "back"}
	default:
		return ""
	}

	var helpBuilder strings.Builder

	for i := 0; i < len(parts); i += 2 {
		if i > 0 {
			helpBuilder.WriteString(theme.Help.Render(" • "))
		}

		key := theme.Selected.Render(fmt.Sprintf(" %s ", parts[i]))
		desc := theme.Help.Render(parts[i+1])
		helpBuilder.WriteString(fmt.Sprintf("%s %s", key, desc))
	}

	return helpBuilder.String()
}

func createFormInputs(name, addr1, addr2 string) []textinput.Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "DNS Name (e.g., Google)"
	nameInput.SetValue(name)
	nameInput.Width = 40
	nameInput.CharLimit = 32

	addr1Input := textinput.New()
	addr1Input.Placeholder = "Primary Address (e.g., 8.8.8.8)"
	addr1Input.SetValue(addr1)
	addr1Input.Width = 40
	addr1Input.Validate = func(s string) error {
		if !utils.IsValidIP(s) {
			return errors.New("invalid IPv4 format")
		}

		return nil
	}

	addr2Input := textinput.New()
	addr2Input.Placeholder = "Secondary Address (Optional)"
	addr2Input.SetValue(addr2)
	addr2Input.Width = 40
	addr2Input.Validate = func(s string) error {
		if s == "" {
			return nil
		}

		if !utils.IsValidIP(s) {
			return errors.New("invalid IPv4 format")
		}

		return nil
	}

	return []textinput.Model{nameInput, addr1Input, addr2Input}
}
