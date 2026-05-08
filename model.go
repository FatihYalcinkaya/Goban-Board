package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RootModel struct {
	columns       []Column
	focusedColumn int
	quitting      bool
	width         int
	height        int
}

func (m RootModel) Init() tea.Cmd { return nil }

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Resize all columns based on terminal width
		numCols := len(m.columns)
		if numCols > 0 {
			colWidth := (msg.Width / numCols) - 6
			colHeight := msg.Height - 8
			for i := range m.columns {
				m.columns[i].list.SetSize(colWidth, colHeight)
			}
		}

	case tea.KeyMsg:
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
		
		case "enter":
			if len(m.columns) > m.focusedColumn+1 {
				selectedTask := m.columns[m.focusedColumn].list.SelectedItem()
				if selectedTask != nil {
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn+1].list.InsertItem(0, selectedTask)
				}
			}
            
		case "n":
			newCol := NewColumn("New Column")
			// Give the new column the same dimensions as others
			if len(m.columns) > 0 {
				w, h := m.columns[0].list.Width(), m.columns[0].list.Height()
				newCol.list.SetSize(w, h)
			}
			m.columns = append(m.columns, newCol)
		}
	}

	var cmd tea.Cmd
	m.columns[m.focusedColumn].list, cmd = m.columns[m.focusedColumn].list.Update(msg)
	return m, cmd
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

	return lipgloss.JoinHorizontal(lipgloss.Top, views...) + 
		"\n\n" + helpStyle.Render(" h/l: switch col • j/k: scroll • n: new col • enter: move right • q: quit")
}
