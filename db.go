package main

import (
	"database/sql"
	"log"

	// Pure Go SQLite driver (no CGO required)
	_ "modernc.org/sqlite"
)

// Global database connection pool
var db *sql.DB

// InitDB initializes the SQLite connection and creates the necessary tables
func InitDB() {
	var err error
	db, err = sql.Open("sqlite", "tasks.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// The fix: Change "NOT nil" to "NOT NULL"
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		status INTEGER NOT NULL
	);`

	_, err = db.Exec(query)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
}

// LoadTasksFromDB retrieves all tasks and populates the model's columns
func LoadTasksFromDB(m *RootModel) {
	rows, err := db.Query("SELECT title, status FROM tasks")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var title string
		var status int
		if err := rows.Scan(&title, &status); err == nil {
			// Ensure the status index matches one of our board columns
			if status >= 0 && status < len(m.columns) {
				// Insert at the end of the list to maintain order
				m.columns[status].list.InsertItem(len(m.columns[status].list.Items()), NewTask(title, ""))
			}
		}
	}
}

// SaveTask persists a new task to the database
func SaveTask(title string, status int) {
	_, err := db.Exec("INSERT INTO tasks (title, description, status) VALUES (?, '', ?)", title, status)
	if err != nil {
		log.Printf("Error saving task: %v", err)
	}
}

// DeleteTask removes a task from the database by its title
func DeleteTask(title string) {
	_, err := db.Exec("DELETE FROM tasks WHERE title = ?", title)
	if err != nil {
		log.Printf("Error deleting task: %v", err)
	}
}

// UpdateTaskStatus updates the column index (status) of a task when moved
func UpdateTaskStatus(title string, newStatus int) {
	_, err := db.Exec("UPDATE tasks SET status = ? WHERE title = ?", newStatus, title)
	if err != nil {
		log.Printf("Error updating task status: %v", err)
	}
}

// RenameTask updates the title of an existing task
func RenameTask(oldTitle string, newTitle string) {
	_, err := db.Exec("UPDATE tasks SET title = ? WHERE title = ?", newTitle, oldTitle)
	if err != nil {
		log.Printf("Error renaming task: %v", err)
	}
}
