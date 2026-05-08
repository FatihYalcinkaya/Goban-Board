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
					// Burası hem yeni ekleme hem de rename sonrası ekleme için çalışır
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

		case "r": // Rename Task
			if len(m.columns[m.focusedColumn].list.Items()) > 0 {
				selected := m.columns[m.focusedColumn].list.SelectedItem().(Task)
				m.input.SetValue(selected.title) // Mevcut başlığı kutuya yaz
				m.state = inputState
				m.input.Focus()
				// Seçili olanı siliyoruz, Enter'a basınca yenisi (günceli) eklenecek
				m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
				return m, nil
			}

		case "d": // Delete Task
			if len(m.columns[m.focusedColumn].list.Items()) > 0 {
				index := m.columns[m.focusedColumn].list.Index()
				m.columns[m.focusedColumn].list.RemoveItem(index)
			}

		case "n": // New Column
			m.columns = append(m.columns, NewColumn("New Column"))
			m.syncDimensions()

		// Movement Logic (Yana Taşıma)
		case "ctrl+l":
			if m.focusedColumn < len(m.columns)-1 {
				selectedItem := m.columns[m.focusedColumn].list.SelectedItem()
				if selectedItem != nil {
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn+1].list.InsertItem(0, selectedItem)
					m.focusedColumn++
				}
			}

		case "ctrl+h":
			if m.focusedColumn > 0 {
				selectedItem := m.columns[m.focusedColumn].list.SelectedItem()
				if selectedItem != nil {
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn-1].list.InsertItem(0, selectedItem)
					m.focusedColumn--
				}
			}

		// Liste İçi Taşıma (Dikey)
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
	return *m
}
func (m *RootModel) syncDimensions() {
	numCols := len(m.columns)
	if numCols == 0 {
		return
	}

	dynWidth := (m.width / numCols) - 5
	if dynWidth < 20 {
		dynWidth = 20
	}

	for i := range m.columns {
		m.columns[i].list.SetSize(dynWidth, m.height-12)

		// Başlık genişliğini sütun içindeki net boşluğa eşitle.
		// -4 birim, sağ ve sol kenarlık paylarını kurtarır.
		m.columns[i].list.Styles.Title = m.columns[i].list.Styles.Title.
			Width(dynWidth - 4).
			MaxWidth(dynWidth - 4) // Taşmasını engelle
	}
}

func (m RootModel) View() string {
	if m.quitting {
		return ""
	}

	var views []string
	numCols := len(m.columns)
	if numCols == 0 {
		return "No columns"
	}

	// Stil genişliğini dinamik hesaplıyoruz
	dynWidth := (m.width / numCols) - 5
	if dynWidth < 20 {
		dynWidth = 20
	}

	for i, col := range m.columns {
		style := columnStyle.Copy().Width(dynWidth)
		if i == m.focusedColumn {
			style = focusedStyle.Copy().Width(dynWidth)
		}
		views = append(views, style.Render(col.list.View()))
	}

	// Sütunları yan yana birleştir
	board := lipgloss.JoinHorizontal(lipgloss.Top, views...)

	// Footer hazırlığı
	var footer string
	if m.state == inputState {
		footer = "\n Edit/Add: " + m.input.View()
	} else {
		footer = helpStyle.Render("\n h/l: move | j/k: nav | ctrl+h/l/j/k: transfer | a: add | r: rename | d: delete | q: quit")
	}

	// Her şeyi dikey birleştir
	fullUI := lipgloss.JoinVertical(lipgloss.Center, board, footer)

	// Tam merkezleme
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		fullUI,
	)
}
