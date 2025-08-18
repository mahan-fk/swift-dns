package tui

import (
	"fmt"
	"strings"
	"swift-dns/internal/config"
	"swift-dns/internal/utils"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func createListView(m *model, title string, items []any, cursor int, renderFunc RenderFunc, selectable bool) {
	headerHeight := lipgloss.Height(viewHeader(*m))
	footerHeight := lipgloss.Height(viewFooter(*m))
	listTitleHeight := lipgloss.Height(title) + 2

	listHeight := max(m.height-
		docStyle.GetVerticalMargins()-
		headerHeight-
		footerHeight-
		listTitleHeight-
		headerAndFooterPadding-
		listTopAndBottomMargin, 0)

	m.list = CustomList{
		Title:        title,
		Items:        items,
		Cursor:       cursor,
		ItemsPerPage: max(listHeight, 1),
		Theme:        &m.currentTheme,
		RenderFunc:   renderFunc,
		Selectable:   selectable,
	}
}

func (m *model) navigateToMainMenu() {
	createListView(m, "Choose an action:", utils.ToAnySlice(m.mainMenuItems), m.mainMenuCursor, renderMainMenuItem, true)
}

func (m *model) navigateToSetDNSList() {
	createListView(m, "Select a DNS provider to apply:", utils.ToAnySlice(m.providers), 0, renderDNSItem, true)
}

func (m *model) navigateToUpdateDNSList(cursor int) {
	createListView(m, "Select a DNS provider to update:", utils.ToAnySlice(m.providers), cursor, renderDNSItem, true)
}

func (m *model) navigateToDeleteDNSList(cursor int) {
	createListView(m, "Select a DNS provider to delete:", utils.ToAnySlice(m.providers), cursor, renderDNSItem, true)
}

func (m *model) navigateToThemeSelection() {
	createListView(m, "Select a theme:", utils.ToAnySlice(m.themesNames), 0, renderThemeItem, true)
}

func (m *model) navigateToInterfaceSelection() {
	createListView(m, "Please select a network interface to manage:", utils.ToAnySlice(m.networkInterfaces), 0, renderInterfaceItem, true)
}

func (m *model) navigateToPingingView() {
	createListView(m, "Ping Results (sorted by fastest):", utils.ToAnySlice(m.pingResults), 0, renderPingResultItem, false)
}

func updateMainMenu(msg tea.KeyMsg, m *model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		selectedItem := m.list.SelectedItem()

		if selectedItem == nil {
			return m, nil
		}

		m.mainMenuCursor = m.list.Cursor
		selectedIndex := m.list.Cursor

		switch selectedIndex {
		case 0:
			m.state = setDNSListView
			m.navigateToSetDNSList()
			return m, nil

		case 1:
			m.state = addDNSView
			m.inputs = createFormInputs("", "", "")
			m.inputFocus = 0
			m.inputs[m.inputFocus].Focus()
			return m, textinput.Blink

		case 2:
			m.state = updateDNSListView
			m.navigateToUpdateDNSList(0)
			return m, nil

		case 3:
			m.state = deleteDNSListView
			m.navigateToDeleteDNSList(0)
			return m, nil

		case 4:
			m.state = loadingView
			m.loadingMessage = "Pinging all DNS servers, please wait..."
			m.pingResults = nil
			return m, tea.Batch(m.spinner.Tick, pingAllCmd(m.providers))

		case 5:
			m.state = loadingView
			m.loadingMessage = "Pinging to find the fastest DNS..."
			m.pingAndSet = true
			m.pingResults = nil
			return m, tea.Batch(m.spinner.Tick, pingAllCmd(m.providers))

		case 6:
			m.state = loadingView
			m.loadingMessage = "Clearing DNS settings..."
			return m, clearDNSCmd(*m)

		case 7:
			m.state = confirmResetView
			return m, nil

		case 8:
			m.state = confirmClearAllView
			return m, nil

		case 9:
			m.state = loadingView
			m.loadingMessage = "Finding active network interfaces..."
			return m, loadInterfacesCmd(m.platform)

		case 10:
			m.state = themeSelectionView
			m.navigateToThemeSelection()
			return m, nil
		}
	}

	updatedList, cmd := m.list.Update(msg)
	m.list = *updatedList
	return m, cmd
}

func handleDNSListUpdate(msg tea.KeyMsg, m *model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.state = mainMenuView
		m.navigateToMainMenu()
		return m, nil

	case "enter":
		selectedItem := m.list.SelectedItem()

		if selectedItem == nil {
			return m, nil
		}

		provider := selectedItem.(config.Provider)

		for i, p := range m.providers {
			if p.Name == provider.Name {
				m.providerToActOn = i
				break
			}
		}

		m.dnsListCursor = m.list.Cursor

		switch m.state {
		case setDNSListView:
			m.state = loadingView
			m.loadingMessage = fmt.Sprintf("Setting DNS to %s...", provider.Name)
			return m, setDNSCmd(*m, provider)

		case updateDNSListView:
			m.state = updateDNSFormView

			if len(provider.Addresses) > 1 {
				m.inputs = createFormInputs(provider.Name, provider.Addresses[0], provider.Addresses[1])
			} else {
				m.inputs = createFormInputs(provider.Name, provider.Addresses[0], "")
			}

			m.inputFocus = 0
			m.inputs[m.inputFocus].Focus()
			return m, textinput.Blink

		case deleteDNSListView:
			m.state = confirmDeleteView
			return m, nil
		}

		return m, nil
	}

	updatedList, cmd := m.list.Update(msg)
	m.list = *updatedList
	return m, cmd
}

func updateFormView(msg tea.KeyMsg, m *model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "tab", "shift+tab", "up", "down":
		m.inputs[m.inputFocus].Blur()

		if msg.String() == "up" || msg.String() == "shift+tab" {
			m.inputFocus--

			if m.inputFocus < 0 {
				m.inputFocus = len(m.inputs) - 1
			}
		} else {
			m.inputFocus++

			if m.inputFocus == len(m.inputs) {
				m.inputFocus = 0
			}
		}

		m.inputs[m.inputFocus].Focus()
		return m, textinput.Blink

	case "esc", "q":
		if m.state == addDNSView {
			m.state = mainMenuView
			m.navigateToMainMenu()
		} else {
			m.state = updateDNSListView
			m.navigateToUpdateDNSList(m.dnsListCursor)
		}

		return m, nil

	case "enter", "ctrl+s":
		name := strings.TrimSpace(m.inputs[0].Value())
		addr1 := strings.TrimSpace(m.inputs[1].Value())
		addr2 := strings.TrimSpace(m.inputs[2].Value())

		if name == "" || addr1 == "" || !utils.IsValidIP(addr1) || (addr2 != "" && !utils.IsValidIP(addr2)) {
			cmd = m.setMessage("Invalid input. Name and a valid Primary DNS are required.", m.currentTheme.Error, 2)
			return m, cmd
		}

		for i, provider := range m.providers {
			if m.state == updateDNSFormView && i == m.providerToActOn {
				continue
			}

			if strings.EqualFold(provider.Name, name) {
				msg := fmt.Sprintf("DNS provider with name '%s' already exists.", name)
				cmd = m.setMessage(msg, m.currentTheme.Error, 2)
				return m, cmd
			}
		}

		newAddresses := []string{addr1}

		if addr2 != "" {
			newAddresses = append(newAddresses, addr2)
		}

		if m.state == addDNSView {
			m.providers = append(m.providers, config.Provider{Name: name, Addresses: newAddresses})
			m.state = mainMenuView
			m.navigateToMainMenu()
		} else {
			m.providers[m.providerToActOn].Name = name
			m.providers[m.providerToActOn].Addresses = newAddresses
			m.state = updateDNSListView
			m.navigateToUpdateDNSList(m.dnsListCursor)
		}

		return m, saveConfigCmd(*m, "DNS provider saved successfully")
	}

	m.inputs[m.inputFocus], cmd = m.inputs[m.inputFocus].Update(msg)
	return m, cmd
}

func updatePingingView(msg tea.KeyMsg, m *model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.state = mainMenuView
		m.navigateToMainMenu()
		return m, nil
	}

	updatedList, cmd := m.list.Update(msg)
	m.list = *updatedList
	return m, cmd
}

func updateConfirmDeleteView(msg tea.KeyMsg, m *model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.providers = append(m.providers[:m.providerToActOn], m.providers[m.providerToActOn+1:]...)
		m.state = deleteDNSListView
		m.navigateToDeleteDNSList(0)
		return m, saveConfigCmd(*m, "DNS provider deleted successfully.")

	case "n", "N", "esc", "q":
		m.state = deleteDNSListView
		m.navigateToDeleteDNSList(m.dnsListCursor)
		return m, nil
	}

	return m, nil
}

func updateConfirmResetView(msg tea.KeyMsg, m *model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.providers = config.GetDefaultProviders()
		m.state = mainMenuView
		m.navigateToMainMenu()
		return m, saveConfigCmd(*m, "DNS list has been reset to default.")

	case "n", "N", "esc", "q":
		m.state = mainMenuView
		return m, nil
	}

	return m, nil
}

func updateConfirmClearAllView(msg tea.KeyMsg, m *model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.providers = []config.Provider{}
		m.state = mainMenuView
		m.navigateToMainMenu()
		return m, saveConfigCmd(*m, "DNS list has been cleared.")

	case "n", "N", "esc", "q":
		m.state = mainMenuView
		return m, nil
	}

	return m, nil
}

func updateThemeSelection(msg tea.KeyMsg, m *model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.state = mainMenuView
		m.navigateToMainMenu()
		return m, nil

	case "enter":
		selectedItem := m.list.SelectedItem()

		if selectedItem == nil {
			return m, nil
		}

		selectedThemeName := selectedItem.(string)
		theme, ok := Themes[selectedThemeName]

		if !ok {
			return m, m.setMessage("Selected theme not found!", m.currentTheme.Error, 2)
		}

		m.currentTheme = theme
		m.currentThemeName = selectedThemeName
		m.spinner.Style = theme.Spinner
		m.state = mainMenuView
		m.navigateToMainMenu()
		return m, saveConfigCmd(*m, fmt.Sprintf("Theme changed to %s", selectedThemeName))
	}

	updatedList, cmd := m.list.Update(msg)
	m.list = *updatedList
	return m, cmd
}

func updateInterfaceSelection(msg tea.KeyMsg, m *model) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		if m.dnsManager == nil {
			return m, tea.Quit
		}

		m.state = mainMenuView
		m.navigateToMainMenu()
		return m, nil

	case "esc":
		if m.dnsManager != nil {
			m.state = mainMenuView
			m.navigateToMainMenu()
			return m, nil
		}

		return m, nil

	case "enter":
		selectedItem := m.list.SelectedItem()

		if selectedItem == nil {
			return m, nil
		}

		iface := selectedItem.(string)
		m.selectedInterface = iface
		dnsManager, err := m.platform.NewDNSManager(m.selectedInterface)

		if err != nil {
			return m, m.setMessage(fmt.Sprintf("Error: %v", err), m.currentTheme.Error, 2)
		}

		if m.dnsManager != nil {
			m.pendingMessage = fmt.Sprintf("Network interface successfully changed to %s", iface)
		}

		m.dnsManager = dnsManager
		m.state = loadingView
		m.loadingMessage = "Checking current DNS status..."
		return m, tea.Batch(saveConfigCmd(*m, ""), checkStatusCmd(*m))
	}

	updatedList, cmd := m.list.Update(msg)
	m.list = *updatedList
	return m, cmd
}
