package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m RootModel) View() string {
	if m.quitting {
		return ""
	}

	numCols := len(m.columns)
	if numCols == 0 {
		return "No columns"
	}

	surfaceStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Background(appBgHex).
		Align(lipgloss.Left, lipgloss.Top)

	if m.state == helpState {
		return surfaceStyle.Render(m.helpView())
	}

	minRequiredWidth := numCols * 25
	if m.width < minRequiredWidth || m.height < 15 {
		errorContent := lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.NewStyle().Foreground(pinkHex).Bold(true).Background(appBgHex).Render("TERMINAL TOO SMALL"),
			lipgloss.NewStyle().Foreground(grayHex).Background(appBgHex).Render("Please enlarge the window"),
		)
		return surfaceStyle.Align(lipgloss.Center, lipgloss.Center).Render(errorContent)
	}

	var views []string
	dynWidth := (m.width / numCols) - 2

	for i := range m.columns {
		m.columns[i].delegate.Styles.NormalTitle = normalTitleStyle

		if i == m.focusedColumn {
			m.columns[i].delegate.Styles.SelectedTitle = focusedTitleStyle
		} else {
			m.columns[i].delegate.Styles.SelectedTitle = selectedTitleStyle
		}
		m.columns[i].list.SetDelegate(m.columns[i].delegate)

		m.columns[i].list.Styles.Title = m.columns[i].list.Styles.Title.Width(dynWidth).MaxWidth(dynWidth)

		style := columnStyle.Width(dynWidth)
		if i == m.focusedColumn {
			style = focusedStyle.Width(dynWidth)
		}
		views = append(views, style.Render(m.columns[i].list.View()))
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, views...)

	footerBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purpleHex).
		Padding(0, 1).
		Background(appBgHex)

	footerContainerStyle := lipgloss.NewStyle().
		Background(appBgHex).
		Width(m.width).
		Align(lipgloss.Center).
		MarginTop(1)

	var footerContent string
	if m.state == confirmState {
		switch m.confirmAction {
		case confirmDeleteTask:
			footerContent = " Delete task? (y/n) "
		case confirmDeleteColumn:
			colName := m.columns[m.focusedColumn].list.Title
			footerContent = fmt.Sprintf(" Delete column '%s' and all its tasks? (y/n) ", colName)
		}
	} else if m.errMsg != "" {
		footerContent = lipgloss.NewStyle().Foreground(pinkHex).Render("ERROR: " + m.errMsg)
	} else if m.state == inputState {
		footerContent = " " + m.input.View()
	} else {
		footerContent = "h/l/j/k: nav | ctrl+h/l/k: transfer | H/L: move col | a: add | A: add col | r: rename | d: delete | R: col-rename | D: col-del | u: undo | ?: help | q: quit"

		contentMaxWidth := m.width - 6
		if contentMaxWidth < 20 {
			contentMaxWidth = 20
		}
		if len(footerContent) > contentMaxWidth {
			var lines []string
			for i := 0; i < len(footerContent); i += contentMaxWidth {
				end := i + contentMaxWidth
				if end > len(footerContent) {
					end = len(footerContent)
				}
				lines = append(lines, footerContent[i:end])
			}
			footerContent = strings.Join(lines, "\n")
		}
	}

	footer := footerContainerStyle.Render(footerBoxStyle.Render(footerContent))

	fullUI := lipgloss.JoinVertical(lipgloss.Left, board, footer)

	return surfaceStyle.Render(fullUI)
}

func (m RootModel) helpView() string {
	help := `Keyboard Shortcuts:

  Navigation:
    h/l              Navigate columns left/right
    j/k              Navigate tasks within a column

  Tasks:
    a                Add new task
    r                Rename selected task
    d                Delete selected task (with confirmation)
    ctrl+h/l         Move task to left/right column
    ctrl+j/k         Reorder task up/down within column

    Columns:
    A                Add new column
    H/L              Move focused column left/right
    R                Rename focused column
    D                Delete focused column (with confirmation)

  Other:
    /                Search/filter tasks in focused column
    u                Undo last deletion
    ?                Toggle this help
    q / ctrl+c       Quit

  Press ? or esc to close this help.`

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purpleHex).
		Padding(1, 2).
		Background(appBgHex).
		Foreground(whiteHex)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Background(appBgHex).
		Align(lipgloss.Center, lipgloss.Center).
		Render(box.Render(help))
}
