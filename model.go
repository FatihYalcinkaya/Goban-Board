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

		// Movement Logic
		case "enter": // Move Right
			if m.focusedColumn < len(m.columns)-1 {
				selectedItem := m.columns[m.focusedColumn].list.SelectedItem()
				if selectedItem != nil {
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn+1].list.InsertItem(0, selectedItem)
				}
			}

		case "backspace": // Move Left
			if m.focusedColumn > 0 {
				selectedItem := m.columns[m.focusedColumn].list.SelectedItem()
				if selectedItem != nil {
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn-1].list.InsertItem(0, selectedItem)
				}
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
	// Terminal boyutu ne olursa olsun sütunlar sabit genişlikte kalsın
	// Ancak yüksekliği terminale göre ayarlamak mantıklıdır
	colHeight := m.height - 10

	for i := range m.columns {
		// fixedColumnWidth stil dosyasından geliyor
		m.columns[i].list.SetSize(fixedColumnWidth, colHeight)
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
		footer = "\n New Task: " + m.input.View() + helpStyle.Render(" (Enter: save, Esc: cancel)")
	} else {
		footer = helpStyle.Render("\n h/l: col • j/k: nav • enter/backspace: move • a: add • d: del • q: quit")
	}

	return board + "\n" + footer
}
