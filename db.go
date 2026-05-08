package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open("sqlite", "tasks.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		log.Fatal("Failed to enable WAL mode:", err)
	}

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

func LoadTasksFromDB(m *RootModel) {
	rows, err := db.Query("SELECT id, title, description, status FROM tasks")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title string
		var desc string
		var status int
		if err := rows.Scan(&id, &title, &desc, &status); err != nil {
			log.Printf("Error scanning task row: %v", err)
			continue
		}
		if status >= 0 && status < len(m.columns) {
			m.columns[status].list.InsertItem(len(m.columns[status].list.Items()), NewTask(id, title, desc))
		}
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating tasks: %v", err)
	}
}

func SaveTask(title string, status int) (int64, error) {
	res, err := db.Exec("INSERT INTO tasks (title, description, status) VALUES (?, '', ?)", title, status)
	if err != nil {
		log.Printf("Error saving task: %v", err)
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func DeleteTask(id int) error {
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		log.Printf("Error deleting task: %v", err)
	}
	return err
}

func UpdateTaskStatus(id int, newStatus int) error {
	_, err := db.Exec("UPDATE tasks SET status = ? WHERE id = ?", newStatus, id)
	if err != nil {
		log.Printf("Error updating task status: %v", err)
	}
	return err
}

func RenameTask(id int, newTitle string) error {
	_, err := db.Exec("UPDATE tasks SET title = ? WHERE id = ?", newTitle, id)
	if err != nil {
		log.Printf("Error renaming task: %v", err)
	}
	return err
}

func UpdateTaskDescription(id int, desc string) error {
	_, err := db.Exec("UPDATE tasks SET description = ? WHERE id = ?", desc, id)
	if err != nil {
		log.Printf("Error updating task description: %v", err)
	}
	return err
}

func DeleteTasksByStatus(status int) error {
	_, err := db.Exec("DELETE FROM tasks WHERE status = ?", status)
	if err != nil {
		log.Printf("Error deleting tasks by status: %v", err)
	}
	return err
}

func UndoDeleteTask(task Task, column int) (int64, error) {
	res, err := db.Exec("INSERT INTO tasks (title, description, status) VALUES (?, ?, ?)", task.title, task.description, column)
	if err != nil {
		log.Printf("Error undoing delete: %v", err)
		return 0, err
	}
	return res.LastInsertId()
}
