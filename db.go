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

	taskQuery := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		status INTEGER NOT NULL
	);`

	if _, err := db.Exec(taskQuery); err != nil {
		return nil, err
	}

	colQuery := `
	CREATE TABLE IF NOT EXISTS columns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		position INTEGER NOT NULL
	);`

	if _, err := db.Exec(colQuery); err != nil {
		return nil, err
	}

	if _, err := db.Exec("ALTER TABLE tasks ADD COLUMN position INTEGER NOT NULL DEFAULT 0"); err != nil {
		// Column may already exist in older databases, ignore
	}

	if _, err := db.Exec("UPDATE tasks SET position = id WHERE position = 0"); err != nil {
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
	rows, err := db.Query("SELECT id, title, status FROM tasks ORDER BY status, position ASC")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title string
		var status int
		if err := rows.Scan(&id, &title, &status); err != nil {
			continue
		}
		if status >= 0 && status < len(m.columns) {
			m.columns[status].list.InsertItem(len(m.columns[status].list.Items()), NewTask(id, title))
		}
	}

	return rows.Err()
}

func SaveTask(db *sql.DB, title string, status int) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE tasks SET position = position + 1 WHERE status = ?", status); err != nil {
		return 0, err
	}
	res, err := tx.Exec("INSERT INTO tasks (title, status, position) VALUES (?, ?, 0)", title, status)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func DeleteTask(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

func UpdateTaskStatus(db *sql.DB, id int, newStatus int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE tasks SET position = position + 1 WHERE status = ?", newStatus); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE tasks SET status = ?, position = 0 WHERE id = ?", newStatus, id); err != nil {
		return err
	}
	return tx.Commit()
}

func RenameTask(db *sql.DB, id int, newTitle string) error {
	_, err := db.Exec("UPDATE tasks SET title = ? WHERE id = ?", newTitle, id)
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

func SwapColumnPositions(db *sql.DB, posA, posB int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE columns SET position = -1 WHERE position = ?", posA); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE columns SET position = ? WHERE position = ?", posA, posB); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE columns SET position = ? WHERE position = -1", posB); err != nil {
		return err
	}

	return tx.Commit()
}

func SwapTaskStatuses(db *sql.DB, statusA, statusB int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE tasks SET status = -1 WHERE status = ?", statusA); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE tasks SET status = ? WHERE status = ?", statusA, statusB); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE tasks SET status = ? WHERE status = -1", statusB); err != nil {
		return err
	}

	return tx.Commit()
}

func SwapTaskPositions(db *sql.DB, idA, idB int) error {
	var posA, posB int
	if err := db.QueryRow("SELECT position FROM tasks WHERE id = ?", idA).Scan(&posA); err != nil {
		return err
	}
	if err := db.QueryRow("SELECT position FROM tasks WHERE id = ?", idB).Scan(&posB); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE tasks SET position = ? WHERE id = ?", posB, idA); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE tasks SET position = ? WHERE id = ?", posA, idB); err != nil {
		return err
	}

	return tx.Commit()
}

func UndoDeleteTask(db *sql.DB, task Task, column int) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE tasks SET position = position + 1 WHERE status = ?", column); err != nil {
		return 0, err
	}
	res, err := tx.Exec("INSERT INTO tasks (title, status, position) VALUES (?, ?, 0)", task.title, column)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
