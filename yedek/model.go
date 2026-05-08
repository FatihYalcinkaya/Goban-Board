package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	defaultState state = iota
	inputState
)

type RootModel struct {
	columns       []Column
	focusedColumn int
	state         state
	input         textinput.Model
	quitting      bool
	width         int
	height        int
	oldTitle      string // Helper to track title changes during rename
}

func (m RootModel) Init() tea.Cmd {
	return nil
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncDimensions()

	case tea.KeyMsg:
		if m.state == inputState {
			switch msg.String() {
			case "enter":
				if m.input.Value() != "" {
					newTitle := m.input.Value()
					// If oldTitle is set, we are in Rename mode
					if m.oldTitle != "" {
						RenameTask(m.oldTitle, newTitle)
						m.columns[m.focusedColumn].list.InsertItem(0, NewTask(newTitle, ""))
						m.oldTitle = "" // Clear after rename
					} else {
						// Normal Add mode
						SaveTask(newTitle, m.focusedColumn)
						m.columns[m.focusedColumn].list.InsertItem(0, NewTask(newTitle, ""))
					}
				}
				return m.resetState(), nil
			case "esc":
				// If canceling a rename, restore the original item
				if m.oldTitle != "" {
					m.columns[m.focusedColumn].list.InsertItem(0, NewTask(m.oldTitle, ""))
					m.oldTitle = ""
				}
				return m.resetState(), nil
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		// --- Normal Mode ---
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		// Column Navigation
		case "left", "h":
			if m.focusedColumn > 0 {
				m.focusedColumn--
			}
		case "right", "l":
			if m.focusedColumn < len(m.columns)-1 {
				m.focusedColumn++
			}

		// Task Actions
		case "a":
			m.state = inputState
			m.input.Focus()
			return m, nil

		case "r":
			if len(m.columns[m.focusedColumn].list.Items()) > 0 {
				selected := m.columns[m.focusedColumn].list.SelectedItem().(Task)
				m.oldTitle = selected.title // Store old title for DB update
				m.input.SetValue(selected.title)
				m.state = inputState
				m.input.Focus()
				// Temporarily remove to avoid duplicates; will be re-added on Enter or Esc
				m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
				return m, nil
			}

		case "d":
			if selectedItem := m.columns[m.focusedColumn].list.SelectedItem(); selectedItem != nil {
				task := selectedItem.(Task)
				DeleteTask(task.title) // Permanent DB removal
				index := m.columns[m.focusedColumn].list.Index()
				m.columns[m.focusedColumn].list.RemoveItem(index)
			}

		case "n":
			m.columns = append(m.columns, NewColumn("New Column"))
			m.syncDimensions()

		// Transfer between columns
		case "ctrl+l":
			if m.focusedColumn < len(m.columns)-1 {
				selectedItem := m.columns[m.focusedColumn].list.SelectedItem()
				if selectedItem != nil {
					task := selectedItem.(Task)
					UpdateTaskStatus(task.title, m.focusedColumn+1) // DB Sync
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn+1].list.InsertItem(0, selectedItem)
					m.focusedColumn++
				}
			}

		case "ctrl+h":
			if m.focusedColumn > 0 {
				selectedItem := m.columns[m.focusedColumn].list.SelectedItem()
				if selectedItem != nil {
					task := selectedItem.(Task)
					UpdateTaskStatus(task.title, m.focusedColumn-1) // DB Sync
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn-1].list.InsertItem(0, selectedItem)
					m.focusedColumn--
				}
			}

		// Vertical movement (No DB change needed since order is handled by Load logic)
		case "ctrl+j":
			curCol := &m.columns[m.focusedColumn]
			index := curCol.list.Index()
			if index < len(curCol.list.Items())-1 {
				selectedItem := curCol.list.SelectedItem()
				curCol.list.RemoveItem(index)
				curCol.list.InsertItem(index+1, selectedItem)
				curCol.list.Select(index + 1)
			}

		case "ctrl+k":
			curCol := &m.columns[m.focusedColumn]
			index := curCol.list.Index()
			if index > 0 {
				selectedItem := curCol.list.SelectedItem()
				curCol.list.RemoveItem(index)
				curCol.list.InsertItem(index-1, selectedItem)
				curCol.list.Select(index - 1)
			}
		}
	}

	var cmd tea.Cmd
	if len(m.columns) > 0 {
		m.columns[m.focusedColumn].list, cmd = m.columns[m.focusedColumn].list.Update(msg)
	}
	return m, cmd
}

func (m *RootModel) resetState() RootModel {
	m.input.Blur()
	m.input.Reset()
	m.state = defaultState
	m.oldTitle = ""
	return *m
}
func (m *RootModel) syncDimensions() {
	numCols := len(m.columns)
	if numCols == 0 {
		return
	}

	// Calculate net width by accounting for border and padding differences
	dynWidth := (m.width / numCols) - 4
	if dynWidth < 20 {
		dynWidth = 20
	}

	for i := range m.columns {
		// Set the physical size of the list
		m.columns[i].list.SetSize(dynWidth, m.height-12)

		// --- CRITICAL: Force styles to width ---
		// This fills the navy gaps on the right with your gray color
		m.columns[i].list.Styles.Title = m.columns[i].list.Styles.Title.
			Width(dynWidth).
			MaxWidth(dynWidth)

		m.columns[i].list.Styles.NoItems = m.columns[i].list.Styles.NoItems.
			Width(dynWidth).
			MaxWidth(dynWidth)
	}
}

func (m RootModel) View() string {
	if m.quitting {
		return ""
	}

	numCols := len(m.columns)
	if numCols == 0 {
		return "No columns"
	}

	// 1. MAIN SURFACE STYLE
	// Seals every cell of the terminal with your color
	surfaceStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Background(appBgHex).
		Align(lipgloss.Left, lipgloss.Top)

	// UI Guard
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
		d := list.NewDefaultDelegate()
		itemStyle := lipgloss.NewStyle().Background(appBgHex)

		d.Styles.NormalTitle = itemStyle.Copy().Foreground(whiteHex)
		d.Styles.NormalDesc = itemStyle.Copy().Foreground(grayHex)
		d.Styles.SelectedTitle = itemStyle.Copy().BorderLeft(false)
		d.Styles.SelectedDesc = itemStyle.Copy().BorderLeft(false)

		if i == m.focusedColumn {
			d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(pinkHex).Bold(true)
			d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(pinkHex)
		} else {
			d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(whiteHex)
			d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(grayHex)
		}
		m.columns[i].list.SetDelegate(d)

		// Force title bar to width
		m.columns[i].list.Styles.Title = m.columns[i].list.Styles.Title.Width(dynWidth).MaxWidth(dynWidth)

		style := columnStyle.Width(dynWidth)
		if i == m.focusedColumn {
			style = focusedStyle.Width(dynWidth)
		}
		views = append(views, style.Render(m.columns[i].list.View()))
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, views...)

	// --- 2. NEW FOOTER DESIGN ---

	// Box Style (Small bordered area)
	footerBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purpleHex). // Matches column border color
		Padding(0, 1).
		Background(appBgHex)

	// Outer Container (Covers the row and centers the box)
	footerContainerStyle := lipgloss.NewStyle().
		Background(appBgHex).
		Width(m.width).
		Align(lipgloss.Center). // Horizontal centering here
		MarginTop(1)

	var footerContent string
	if m.state == inputState {
		footerContent = " Edit/Add: " + m.input.View()
	} else {
		footerContent = "h/l/j/k: move | ctrl+h/l/j/k: transfer | a: add | r: rename | d: delete | q: quit"
	}

	footer := footerContainerStyle.Render(footerBoxStyle.Render(footerContent))

	// 3. FINAL MERGE
	// Merge by left-aligning content (Right-side gap will be covered by surfaceStyle)
	fullUI := lipgloss.JoinVertical(lipgloss.Left, board, footer)

	return surfaceStyle.Render(fullUI)
}
