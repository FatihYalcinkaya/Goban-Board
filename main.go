package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	// Import the pure Go SQLite driver
	_ "modernc.org/sqlite"
)

func main() {
	// Initialize the database and create tables if they don't exist
	InitDB()

	ti := textinput.New()
	ti.Placeholder = "Enter task title..."
	ti.CharLimit = 50
	ti.Width = 30

	// Initialize the kanban columns
	backlog := NewColumn("BACKLOG")
	todo := NewColumn("TO DO")
	inProgress := NewColumn("IN PROGRESS")
	done := NewColumn("DONE")

	m := RootModel{
		columns: []Column{
			backlog,
			todo,
			inProgress,
			done,
		},
		input:         ti,
		focusedColumn: 1,
	}

	// Load existing tasks from the database into the model
	LoadTasksFromDB(&m)

	// Run the Bubble Tea program with AltScreen enabled
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
