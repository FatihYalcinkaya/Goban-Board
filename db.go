package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return nil, err
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
		return nil, err
	}

	colQuery := `
	CREATE TABLE IF NOT EXISTS columns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		position INTEGER NOT NULL
	);`

	_, err = db.Exec(colQuery)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func LoadColumnsFromDB(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT name FROM columns ORDER BY position ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

func SaveColumn(db *sql.DB, name string, position int) error {
	_, err := db.Exec("INSERT INTO columns (name, position) VALUES (?, ?)", name, position)
	return err
}

func DeleteColumnByPosition(db *sql.DB, position int) error {
	_, err := db.Exec("DELETE FROM columns WHERE position = ?", position)
	return err
}

func ShiftColumnPositions(db *sql.DB, fromPosition int) error {
	_, err := db.Exec("UPDATE columns SET position = position - 1 WHERE position > ?", fromPosition)
	return err
}

func RenameColumn(db *sql.DB, position int, newName string) error {
	_, err := db.Exec("UPDATE columns SET name = ? WHERE position = ?", newName, position)
	return err
}

func LoadTasksFromDB(db *sql.DB, m *RootModel) error {
	rows, err := db.Query("SELECT id, title, description, status FROM tasks")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title string
		var desc string
		var status int
		if err := rows.Scan(&id, &title, &desc, &status); err != nil {
			continue
		}
		if status >= 0 && status < len(m.columns) {
			m.columns[status].list.InsertItem(len(m.columns[status].list.Items()), NewTask(id, title, desc))
		}
	}

	return rows.Err()
}

func SaveTask(db *sql.DB, title string, status int) (int64, error) {
	res, err := db.Exec("INSERT INTO tasks (title, description, status) VALUES (?, '', ?)", title, status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func DeleteTask(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

func UpdateTaskStatus(db *sql.DB, id int, newStatus int) error {
	_, err := db.Exec("UPDATE tasks SET status = ? WHERE id = ?", newStatus, id)
	return err
}

func RenameTask(db *sql.DB, id int, newTitle string) error {
	_, err := db.Exec("UPDATE tasks SET title = ? WHERE id = ?", newTitle, id)
	return err
}

func UpdateTaskDescription(db *sql.DB, id int, desc string) error {
	_, err := db.Exec("UPDATE tasks SET description = ? WHERE id = ?", desc, id)
	return err
}

func DeleteTasksByStatus(db *sql.DB, status int) error {
	_, err := db.Exec("DELETE FROM tasks WHERE status = ?", status)
	return err
}

func ShiftTaskStatuses(db *sql.DB, fromColIdx int) error {
	_, err := db.Exec("UPDATE tasks SET status = status - 1 WHERE status > ?", fromColIdx)
	return err
}

func UndoDeleteTask(db *sql.DB, task Task, column int) (int64, error) {
	res, err := db.Exec("INSERT INTO tasks (title, description, status) VALUES (?, ?, ?)", task.title, task.description, column)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
