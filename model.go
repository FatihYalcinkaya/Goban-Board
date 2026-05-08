package main

import (
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
		// Logic when typing a new task (Input Mode)
		if m.state == inputState {
			switch msg.String() {
			case "enter":
				if m.input.Value() != "" {
					m.columns[m.focusedColumn].list.InsertItem(0, NewTask(m.input.Value(), ""))
				}
				return m.resetState(), nil
			case "esc":
				return m.resetState(), nil
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		// Normal Mode Keybindings
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "left", "h":
			if m.focusedColumn > 0 {
				m.focusedColumn--
			}
		case "right", "l":
			if m.focusedColumn < len(m.columns)-1 {
				m.focusedColumn++
			}

		case "a": // Add Task
			m.state = inputState
			m.input.Focus()
			return m, nil

		case "d": // Delete Task
			if len(m.columns[m.focusedColumn].list.Items()) > 0 {
				index := m.columns[m.focusedColumn].list.Index()
				m.columns[m.focusedColumn].list.RemoveItem(index)
			}

		case "n": // New Column
			m.columns = append(m.columns, NewColumn("New Column"))
			m.syncDimensions()

		case "enter": // Move Task to the Right
			if len(m.columns) > m.focusedColumn+1 {
				selectedTask := m.columns[m.focusedColumn].list.SelectedItem()
				if selectedTask != nil {
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn+1].list.InsertItem(0, selectedTask)
				}
			}
		}
	}

	// Important: Pass keys (like j/k) to the focused list for navigation
	var cmd tea.Cmd
	if len(m.columns) > 0 {
		m.columns[m.focusedColumn].list, cmd = m.columns[m.focusedColumn].list.Update(msg)
	}
	return m, cmd
}

// resetState exits the input mode and clears the text box
func (m *RootModel) resetState() RootModel {
	m.input.Blur()
	m.input.Reset()
	m.state = defaultState
	return *m
}

// syncDimensions recalculates column sizes based on terminal width/height
func (m *RootModel) syncDimensions() {
	numCols := len(m.columns)
	if numCols > 0 {
		// Calculate space per column subtracting borders and padding
		colWidth := (m.width / numCols) - 6
		colHeight := m.height - 10
		for i := range m.columns {
			m.columns[i].list.SetSize(colWidth, colHeight)
		}
	}
}

func (m RootModel) View() string {
	if m.quitting {
		return ""
	}

	var views []string
	for i, col := range m.columns {
		style := columnStyle
		if i == m.focusedColumn {
			style = focusedStyle
		}
		views = append(views, style.Render(col.list.View()))
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, views...)

	// Footer logic
	footer := helpStyle.Render("\n h/l: switch col • j/k: nav • a: add • d: delete • enter: move right • q: quit")
	if m.state == inputState {
		footer = "\n New Task: " + m.input.View() + helpStyle.Render(" (Enter to save, Esc to cancel)")
	}

	return board + "\n" + footer
}
