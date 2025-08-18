package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle               = lipgloss.NewStyle().Margin(1, 2)
	headerAndFooterPadding = 0
	listTopAndBottomMargin = 2
)

func (m *model) Init() tea.Cmd {
	if m.dnsManager == nil {
		return tea.Batch(m.spinner.Tick, loadInterfacesCmd(m.platform))
	}

	return tea.Batch(m.spinner.Tick, checkStatusCmd(*m))
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height

		switch m.state {
		case mainMenuView:
			m.navigateToMainMenu()
		case setDNSListView:
			m.navigateToSetDNSList()
		case updateDNSListView:
			m.navigateToUpdateDNSList(m.dnsListCursor)
		case deleteDNSListView:
			m.navigateToDeleteDNSList(m.dnsListCursor)
		case pingingView:
			m.navigateToPingingView()
		case themeSelectionView:
			m.navigateToThemeSelection()
		case interfaceSelectionView:
			m.navigateToInterfaceSelection()
		}

		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || (msg.String() == "q" && m.state == mainMenuView) {
			return m, tea.Quit
		}

		switch m.state {
		case interfaceSelectionView:
			return updateInterfaceSelection(msg, m)
		case mainMenuView:
			return updateMainMenu(msg, m)
		case setDNSListView, updateDNSListView, deleteDNSListView:
			return handleDNSListUpdate(msg, m)
		case addDNSView, updateDNSFormView:
			return updateFormView(msg, m)
		case pingingView:
			return updatePingingView(msg, m)
		case confirmDeleteView:
			return updateConfirmDeleteView(msg, m)
		case confirmResetView:
			return updateConfirmResetView(msg, m)
		case confirmClearAllView:
			return updateConfirmClearAllView(msg, m)
		case themeSelectionView:
			return updateThemeSelection(msg, m)
		}

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case interfacesLoadedMsg:
		m.networkInterfaces = msg
		m.state = interfaceSelectionView
		m.navigateToInterfaceSelection()
		return m, nil

	case statusMsg:
		m.status = string(msg)
		m.state = mainMenuView
		m.navigateToMainMenu()

		if m.pendingMessage != "" {
			cmd = m.setMessage(m.pendingMessage, m.currentTheme.Success, 2)
			m.pendingMessage = ""
			return m, cmd
		}

		return m, nil

	case dnsSetMsg:
		provider := msg.provider
		m.status = fmt.Sprintf("%s [%s]", provider.Name, strings.Join(provider.Addresses, " | "))
		m.state = mainMenuView
		m.navigateToMainMenu()
		successMsg := fmt.Sprintf("DNS successfully set to %s", provider.Name)
		cmd = m.setMessage(successMsg, m.currentTheme.Success, 2)
		return m, cmd

	case listUpdatedMsg:
		if m.state == loadingView {
			return m, nil
		}

		originalMsg := string(msg)

		if originalMsg != "" {
			cmd = m.setMessage(originalMsg, m.currentTheme.Success, 2)
			cmds = append(cmds, cmd)
		}

		if m.state != deleteDNSListView && m.state != updateDNSListView {
			m.state = mainMenuView
			m.navigateToMainMenu()
		}

		return m, tea.Batch(cmds...)

	case dnsClearedMsg:
		m.status = "DHCP [Automatic]"
		m.state = mainMenuView
		m.navigateToMainMenu()
		cmd = m.setMessage(string(msg), m.currentTheme.Success, 2)
		return m, cmd

	case errorMsg:
		m.state = mainMenuView
		m.navigateToMainMenu()
		m.pingAndSet = false
		errorText := fmt.Sprintf("Error: %v", msg.err)
		cmd = m.setMessage(errorText, m.currentTheme.Error, 4)
		return m, cmd

	case pingResultsMsg:
		m.pingResults = msg

		if m.pingAndSet {
			m.pingAndSet = false

			if len(m.pingResults) > 1 && !m.pingResults[0].Result.TimedOut {
				fastestProvider := m.pingResults[0].Provider
				m.state = loadingView
				m.loadingMessage = fmt.Sprintf("Setting DNS to fastest provider: %s", fastestProvider.Name)
				return m, setDNSCmd(*m, fastestProvider)
			} else {
				m.state = mainMenuView
				m.navigateToMainMenu()
				cmd = m.setMessage("All DNS pings timed out.", m.currentTheme.Error, 3)
				return m, cmd
			}
		}

		m.state = pingingView
		m.navigateToPingingView()
		return m, nil

	case clearMessageMsg:
		if msg.id == m.messageID {
			m.message = ""
		}

		return m, nil
	}

	return m, cmd
}

func (m *model) View() string {
	if m.state == loadingView {
		return docStyle.Render(viewLoading(*m))
	}

	var body string
	switch m.state {
	case mainMenuView, setDNSListView, updateDNSListView, deleteDNSListView, pingingView, interfaceSelectionView, themeSelectionView:
		body = m.list.View()
	case addDNSView, updateDNSFormView:
		body = viewForm(*m)
	case confirmDeleteView:
		body = viewConfirm(*m, "delete", m.providers[m.providerToActOn].Name)
	case confirmResetView:
		body = viewConfirm(*m, "reset", "the entire DNS list")
	case confirmClearAllView:
		body = viewConfirm(*m, "clear", "the entire DNS list")
	}

	header := viewHeader(*m)
	footer := viewFooter(*m)

	headerHeight := lipgloss.Height(header)
	availableHeight := m.height - docStyle.GetVerticalMargins()
	footerHeight := lipgloss.Height(footer)
	bodyHeight := max(availableHeight-headerHeight-footerHeight-headerAndFooterPadding, 0)
	body = lipgloss.NewStyle().Height(bodyHeight).Render(body)
	finalView := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

	return docStyle.Render(finalView)
}
