package tui

import (
	"sort"
	"swift-dns/internal/config"
	"swift-dns/internal/dns"
	"swift-dns/internal/ping"
	"swift-dns/internal/platform"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewState int

const (
	loadingView viewState = iota
	mainMenuView
	setDNSListView
	updateDNSListView
	deleteDNSListView
	addDNSView
	updateDNSFormView
	pingingView
	interfaceSelectionView
	confirmDeleteView
	confirmResetView
	confirmClearAllView
	themeSelectionView
)

type statusMsg string
type pingResultsMsg []ping.JobResult
type errorMsg struct{ err error }
type dnsSetMsg struct{ provider config.Provider }
type dnsClearedMsg string
type listUpdatedMsg string
type interfacesLoadedMsg []string
type clearMessageMsg struct{ id int }

type model struct {
	dnsManager        dns.Manager
	platform          platform.Platform
	configPath        string
	state             viewState
	providers         []config.Provider
	mainMenuCursor    int
	dnsListCursor     int
	status            string
	pingResults       []ping.JobResult
	message           string
	messageID         int
	loadingMessage    string
	pendingMessage    string
	spinner           spinner.Model
	inputs            []textinput.Model
	inputFocus        int
	providerToActOn   int
	themesNames       []string
	currentTheme      Theme
	currentThemeName  string
	mainMenuItems     []string
	networkInterfaces []string
	selectedInterface string
	pingAndSet        bool
	list              CustomList
	height            int
}

func (m *model) setMessage(msg string, style lipgloss.Style, seconds int) tea.Cmd {
	m.message = style.Render(msg)
	m.messageID++
	return clearMessageAfterCmd(m.messageID, time.Duration(seconds)*time.Second)
}

func InitialModel(p platform.Platform, configPath string, dm dns.Manager, cfg *config.Config) model {
	loadedThemeName := cfg.ThemeName
	theme, ok := Themes[loadedThemeName]

	if !ok {
		loadedThemeName = "Default"
		theme = Themes[loadedThemeName]
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.Spinner

	var themeNames []string

	for name := range Themes {
		themeNames = append(themeNames, name)
	}

	sort.Strings(themeNames)

	mainMenuItems := []string{
		"Set a DNS",
		"Add a new DNS",
		"Edit a DNS",
		"Delete a DNS",
		"Ping all DNS servers",
		"Set fastest DNS",
		"Revert to DHCP (Automatic)",
		"Restore default DNS list",
		"Clear DNS List",
		"Change network interface",
		"Customize appearance",
	}

	loadingMsg := "Checking current DNS status..."

	if dm == nil {
		loadingMsg = "Finding active network interfaces..."
	}

	return model{
		platform:          p,
		dnsManager:        dm,
		configPath:        configPath,
		providers:         cfg.Providers,
		state:             loadingView,
		spinner:           s,
		themesNames:       themeNames,
		currentTheme:      theme,
		currentThemeName:  loadedThemeName,
		mainMenuItems:     mainMenuItems,
		selectedInterface: cfg.SelectedInterface,
		loadingMessage:    loadingMsg,
		pingAndSet:        false,
		status:            "---",
	}
}
