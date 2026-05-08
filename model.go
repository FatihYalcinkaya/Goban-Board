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
		// --- Input Mode Logic ---
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

		// --- Normal Mode Logic ---
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

		// Movement Logic (Odağın kartla beraber kayması eklendi)
		case "ctrl+l": // Move Right
			if m.focusedColumn < len(m.columns)-1 {
				selectedItem := m.columns[m.focusedColumn].list.SelectedItem()
				if selectedItem != nil {
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn+1].list.InsertItem(0, selectedItem)
					// Fokus kartla beraber sağa kayar
					m.focusedColumn++
				}
			}

		case "ctrl+h": // Move Left
			if m.focusedColumn > 0 {
				selectedItem := m.columns[m.focusedColumn].list.SelectedItem()
				if selectedItem != nil {
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn-1].list.InsertItem(0, selectedItem)
					// Fokus kartla beraber sola kayar
					m.focusedColumn--
				}
			}
		case "ctrl+j": // Aşağı taşı
			curCol := &m.columns[m.focusedColumn]
			index := curCol.list.Index()
			items := curCol.list.Items()
			if index < len(items)-1 {
				selectedItem := curCol.list.SelectedItem()
				curCol.list.RemoveItem(index)
				curCol.list.InsertItem(index+1, selectedItem)
				curCol.list.Select(index + 1) // İmleci de kaydır ki takibi kolay olsun
			}

		case "ctrl+k": // Yukarı taşı
			curCol := &m.columns[m.focusedColumn]
			index := curCol.list.Index()
			if index > 0 {
				selectedItem := curCol.list.SelectedItem()
				curCol.list.RemoveItem(index)
				curCol.list.InsertItem(index-1, selectedItem)
				curCol.list.Select(index - 1) // İmleci de kaydır
			}
		}
	}

	// Delegate messages to the active list (handles j/k navigation)
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
	return *m
}

func (m *RootModel) syncDimensions() {
	// Yüksekliği terminale göre ayarlıyoruz
	colHeight := m.height - 10

	for i := range m.columns {
		// Listeye stil genişliğinden biraz daha az yer veriyoruz
		// (kenarlık ve padding payı için fixedColumnWidth-4 idealdir)
		m.columns[i].list.SetSize(fixedColumnWidth-4, colHeight)
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

	var footer string
	if m.state == inputState {
		footer = "\n Add Task: " + m.input.View() + helpStyle.Render(" (Enter: save, Esc: cancel)")
	} else {
		footer = helpStyle.Render("\n h: move left l: move right ctrl+h: move task left ctrl+l: move task right d:delete task q:quit ")
	}

	return board + "\n" + footer
}
