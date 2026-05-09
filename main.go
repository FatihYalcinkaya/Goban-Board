package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	_ "modernc.org/sqlite"
)

func main() {
	dbPath := os.Getenv("KANBAN_DB_PATH")
	if dbPath == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			fmt.Printf("Failed to get config directory: %v\n", err)
			return
		}
		dbDir := filepath.Join(configDir, "goban board")
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			fmt.Printf("Failed to create config directory: %v\n", err)
			return
		}
		dbPath = filepath.Join(dbDir, "tasks.db")
	}

	db, err := InitDB(dbPath)
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}
	defer db.Close()

	ti := textinput.New()
	ti.Placeholder = "Enter task title..."
	ti.CharLimit = 50
	ti.Width = 30

	colNames, err := LoadColumnsFromDB(db)
	if err != nil {
		fmt.Printf("Failed to load columns: %v\n", err)
		return
	}

	var columns []Column
	if len(colNames) == 0 {
		defaults := []string{"BACKLOG", "TO DO", "IN PROGRESS", "DONE"}
		for i, name := range defaults {
			columns = append(columns, NewColumn(name))
			if err := SaveColumn(db, name, i); err != nil {
				fmt.Printf("Failed to save default column: %v\n", err)
				return
			}
		}
	} else {
		for _, name := range colNames {
			columns = append(columns, NewColumn(name))
		}
	}

	focusedColumn := 1
	if focusedColumn >= len(columns) {
		focusedColumn = len(columns) - 1
	}

	m := RootModel{
		columns:         columns,
		input:           ti,
		focusedColumn:   focusedColumn,
		db:              db,
		columnRenameIdx: -1,
	}

	if err := LoadTasksFromDB(db, &m); err != nil {
		fmt.Printf("Failed to load tasks: %v\n", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		return
	}
}
