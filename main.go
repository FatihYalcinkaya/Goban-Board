package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	todo := NewColumn("TODO")
	todo.list.InsertItem(0, NewTask("Build TUI", "Use Go and Bubble Tea"))
	todo.list.InsertItem(1, NewTask("Configure Neovim", "Update Lua plugins"))

	inProg := NewColumn("IN PROGRESS")
	inProg.list.InsertItem(0, NewTask("Refactoring", "Fixing unused imports"))

	done := NewColumn("DONE")

	m := RootModel{
		columns: []Column{todo, inProg, done},
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
