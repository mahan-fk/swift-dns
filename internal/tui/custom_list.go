package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RenderFunc func(item any, isSelected bool, theme *Theme) string

type CustomList struct {
	Title        string
	Items        []any
	Cursor       int
	ItemsPerPage int
	Theme        *Theme
	RenderFunc   RenderFunc
	Selectable   bool
}

func (list *CustomList) Update(msg tea.Msg) (*CustomList, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case list.Selectable && key.Matches(msg, key.NewBinding(key.WithKeys("up"))):
			if list.Cursor > 0 {
				list.Cursor--
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down"))):
			if list.Selectable && list.Cursor < len(list.Items)-1 {
				list.Cursor++
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("left"))):
			if list.ItemsPerPage > 0 {
				list.Cursor -= list.ItemsPerPage

				if list.Cursor < 0 {
					list.Cursor = 0
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("right"))):
			if list.ItemsPerPage > 0 {
				list.Cursor += list.ItemsPerPage

				if list.Cursor >= len(list.Items) {
					list.Cursor = len(list.Items) - 1
				}
			}
		}
	}

	return list, nil
}

func (list *CustomList) View() string {
	if list.RenderFunc == nil {
		return "Error: RenderFunc is not set for CustomList"
	}

	var b strings.Builder

	b.WriteString(list.Title + "\n\n")

	if len(list.Items) == 0 {
		b.WriteString("  No items to display.\n")
		return b.String()
	}

	currentPage := 0
	start := 0
	end := len(list.Items)

	if list.ItemsPerPage > 0 {
		currentPage = list.Cursor / list.ItemsPerPage
		start = currentPage * list.ItemsPerPage
		end = min(start+list.ItemsPerPage, len(list.Items))
	}

	visibleItems := list.Items[start:end]

	for i, item := range visibleItems {
		cursorInPage := 0

		if list.ItemsPerPage > 0 {
			cursorInPage = list.Cursor % list.ItemsPerPage
		}

		isSelected := (i == cursorInPage)

		b.WriteString(list.RenderFunc(item, isSelected, list.Theme))
		b.WriteString("\n")
	}

	numVisible := len(visibleItems)

	if list.ItemsPerPage > 0 && numVisible < list.ItemsPerPage {
		b.WriteString(strings.Repeat("\n", list.ItemsPerPage-numVisible))
	}

	pagination := list.paginationView()

	b.WriteString(pagination)

	if pagination != "" {
		b.WriteString("\n")
	}

	return b.String()
}

func (list *CustomList) paginationView() string {
	if list.ItemsPerPage <= 0 || len(list.Items) <= list.ItemsPerPage {
		return ""
	}

	totalPages := int(math.Ceil(float64(len(list.Items)) / float64(list.ItemsPerPage)))
	currentPage := list.Cursor / list.ItemsPerPage

	dotStyle := lipgloss.NewStyle().Foreground(list.Theme.Help.GetForeground())
	activeDotStyle := lipgloss.NewStyle().Foreground(list.Theme.Selected.GetForeground())

	var dots strings.Builder

	for i := range totalPages {
		if i == currentPage {
			dots.WriteString(activeDotStyle.Render("•"))
		} else {
			dots.WriteString(dotStyle.Render("•"))
		}
	}

	return lipgloss.NewStyle().MarginTop(1).Render(dots.String())
}

func (list *CustomList) SelectedItem() any {
	if !list.Selectable || len(list.Items) == 0 || list.Cursor < 0 || list.Cursor >= len(list.Items) {
		return nil
	}

	return list.Items[list.Cursor]
}
