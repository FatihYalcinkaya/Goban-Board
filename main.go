package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// 1. Initialize the text input for adding tasks
	ti := textinput.New()
	ti.Placeholder = "Enter task title..."
	ti.CharLimit = 50
	ti.Width = 30

	// 2. Define the default columns in the requested order
	backlog := NewColumn("BACKLOG")
	todo := NewColumn("TO DO")
	inProgress := NewColumn("IN PROGRESS")
	done := NewColumn("DONE")

	// 4. Create the RootModel with the ordered columns
	m := RootModel{
		columns: []Column{
			backlog,
			todo,
			inProgress,
			done,
		},
		input: ti,
	}

	// 5. Run the program using the AltScreen to keep the terminal clean
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
