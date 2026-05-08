package main

import (
	"fmt"
	"os"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	ti := textinput.New()
	ti.Placeholder = "Task Title..."
	ti.Focus()

	todo := NewColumn("TODO")
	todo.list.InsertItem(0, NewTask("Buy Coffee", "Need fuel"))
	
	m := RootModel{
		columns: []Column{todo, NewColumn("DONE")},
		input:   ti,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
