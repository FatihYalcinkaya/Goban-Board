package main

import (
	"database/sql"
	"fmt"
	"testing"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func newTestBoard() *RootModel {
	return &RootModel{
		columns: []Column{
			NewColumn("BACKLOG"),
			NewColumn("TO DO"),
			NewColumn("IN PROGRESS"),
			NewColumn("DONE"),
		},
	}
}

func TestSaveAndLoadTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	m := newTestBoard()
	id, err := SaveTask(db, "Test task", 0)
	if err != nil {
		t.Fatalf("SaveTask failed: %v", err)
	}
	if id == 0 {
		t.Fatal("Expected non-zero ID")
	}

	if err := LoadTasksFromDB(db, m); err != nil {
		t.Fatalf("LoadTasksFromDB failed: %v", err)
	}

	items := m.columns[0].list.Items()
	if len(items) != 1 {
		t.Fatalf("Expected 1 item in column 0, got %d", len(items))
	}
	task, ok := items[0].(Task)
	if !ok {
		t.Fatal("Expected Task type")
	}
	if task.title != "Test task" {
		t.Fatalf("Expected title 'Test task', got '%s'", task.title)
	}
	if task.description != "" {
		t.Fatalf("Expected empty description, got '%s'", task.description)
	}
}

func TestSaveTaskWithDescription(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id, err := SaveTask(db, "Task with desc", 0)
	if err != nil {
		t.Fatalf("SaveTask failed: %v", err)
	}

	if err := UpdateTaskDescription(db, int(id), "A description"); err != nil {
		t.Fatalf("UpdateTaskDescription failed: %v", err)
	}

	m := newTestBoard()
	if err := LoadTasksFromDB(db, m); err != nil {
		t.Fatalf("LoadTasksFromDB failed: %v", err)
	}

	task := m.columns[0].list.Items()[0].(Task)
	if task.description != "A description" {
		t.Fatalf("Expected 'A description', got '%s'", task.description)
	}
}

func TestDeleteTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id, err := SaveTask(db, "To delete", 0)
	if err != nil {
		t.Fatalf("SaveTask failed: %v", err)
	}

	if err := DeleteTask(db, int(id)); err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	m := &RootModel{columns: []Column{NewColumn("BACKLOG")}}
	if err := LoadTasksFromDB(db, m); err != nil {
		t.Fatalf("LoadTasksFromDB failed: %v", err)
	}
	if len(m.columns[0].list.Items()) != 0 {
		t.Fatal("Expected 0 items after deletion")
	}
}

func TestRenameTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id, err := SaveTask(db, "Original", 0)
	if err != nil {
		t.Fatalf("SaveTask failed: %v", err)
	}

	if err := RenameTask(db, int(id), "Renamed"); err != nil {
		t.Fatalf("RenameTask failed: %v", err)
	}

	m := &RootModel{columns: []Column{NewColumn("BACKLOG")}}
	if err := LoadTasksFromDB(db, m); err != nil {
		t.Fatalf("LoadTasksFromDB failed: %v", err)
	}
	task := m.columns[0].list.Items()[0].(Task)
	if task.title != "Renamed" {
		t.Fatalf("Expected title 'Renamed', got '%s'", task.title)
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id, err := SaveTask(db, "Moving", 0)
	if err != nil {
		t.Fatalf("SaveTask failed: %v", err)
	}

	if err := UpdateTaskStatus(db, int(id), 2); err != nil {
		t.Fatalf("UpdateTaskStatus failed: %v", err)
	}

	m := newTestBoard()
	if err := LoadTasksFromDB(db, m); err != nil {
		t.Fatalf("LoadTasksFromDB failed: %v", err)
	}
	if len(m.columns[0].list.Items()) != 0 {
		t.Fatal("Expected 0 items in column 0")
	}
	if len(m.columns[2].list.Items()) != 1 {
		t.Fatalf("Expected 1 item in column 2, got %d", len(m.columns[2].list.Items()))
	}
}

func TestDeleteTasksByStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	SaveTask(db, "Task A", 0)
	SaveTask(db, "Task B", 1)
	SaveTask(db, "Task C", 0)

	if err := DeleteTasksByStatus(db, 0); err != nil {
		t.Fatalf("DeleteTasksByStatus failed: %v", err)
	}

	m := &RootModel{
		columns: []Column{NewColumn("BACKLOG"), NewColumn("TO DO")},
	}
	if err := LoadTasksFromDB(db, m); err != nil {
		t.Fatalf("LoadTasksFromDB failed: %v", err)
	}
	if len(m.columns[0].list.Items()) != 0 {
		t.Fatal("Expected 0 items in column 0")
	}
	if len(m.columns[1].list.Items()) != 1 {
		t.Fatalf("Expected 1 item in column 1, got %d", len(m.columns[1].list.Items()))
	}
}

func TestShiftTaskStatuses(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	SaveTask(db, "Task A", 0)
	SaveTask(db, "Task B", 1)
	SaveTask(db, "Task C", 2)

	// Simulate deleting column 0: delete tasks with that status, then shift remaining
	if err := DeleteTasksByStatus(db, 0); err != nil {
		t.Fatalf("DeleteTasksByStatus failed: %v", err)
	}
	if err := ShiftTaskStatuses(db, 0); err != nil {
		t.Fatalf("ShiftTaskStatuses failed: %v", err)
	}

	m := &RootModel{
		columns: []Column{NewColumn("BACKLOG"), NewColumn("IN PROGRESS"), NewColumn("DONE")},
	}
	if err := LoadTasksFromDB(db, m); err != nil {
		t.Fatalf("LoadTasksFromDB failed: %v", err)
	}
	if len(m.columns[0].list.Items()) != 1 {
		t.Fatalf("Expected 1 item in column 0 (Task B shifted from 1), got %d", len(m.columns[0].list.Items()))
	}
	if len(m.columns[1].list.Items()) != 1 {
		t.Fatalf("Expected 1 item in column 1 (Task C shifted from 2), got %d", len(m.columns[1].list.Items()))
	}
	if len(m.columns[2].list.Items()) != 0 {
		t.Fatalf("Expected 0 items in column 2, got %d", len(m.columns[2].list.Items()))
	}
}

func TestUndoDeleteTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	id1, _ := SaveTask(db, "Task 1", 0)
	SaveTask(db, "Task 2", 1)

	if err := DeleteTask(db, int(id1)); err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	newID, err := UndoDeleteTask(db, Task{id: int(id1), title: "Task 1", description: ""}, 0)
	if err != nil {
		t.Fatalf("UndoDeleteTask failed: %v", err)
	}
	if newID == 0 {
		t.Fatal("Expected non-zero new ID")
	}

	m := &RootModel{
		columns: []Column{NewColumn("BACKLOG"), NewColumn("TO DO")},
	}
	if err := LoadTasksFromDB(db, m); err != nil {
		t.Fatalf("LoadTasksFromDB failed: %v", err)
	}
	if len(m.columns[0].list.Items()) != 1 {
		t.Fatalf("Expected 1 item in column 0, got %d", len(m.columns[0].list.Items()))
	}
	if len(m.columns[1].list.Items()) != 1 {
		t.Fatalf("Expected 1 item in column 1, got %d", len(m.columns[1].list.Items()))
	}
}

func TestSaveMultipleTasks(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	for i := 0; i < 10; i++ {
		if _, err := SaveTask(db, fmt.Sprintf("Task %d", i), i%4); err != nil {
			t.Fatalf("SaveTask %q failed: %v", fmt.Sprintf("Task %d", i), err)
		}
	}

	m := newTestBoard()
	if err := LoadTasksFromDB(db, m); err != nil {
		t.Fatalf("LoadTasksFromDB failed: %v", err)
	}

	total := 0
	for _, col := range m.columns {
		total += len(col.list.Items())
	}
	if total != 10 {
		t.Fatalf("Expected 10 total items, got %d", total)
	}
}

func TestLoadTasksFromEmptyDB(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	m := &RootModel{columns: []Column{NewColumn("BACKLOG")}}
	if err := LoadTasksFromDB(db, m); err != nil {
		t.Fatalf("LoadTasksFromDB on empty DB failed: %v", err)
	}
	if len(m.columns[0].list.Items()) != 0 {
		t.Fatal("Expected 0 items from empty DB")
	}
}

func TestDeleteNonexistentTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	if err := DeleteTask(db, 999); err != nil {
		t.Fatalf("DeleteTask on nonexistent task failed: %v", err)
	}
}

func TestInitDBInvalidPath(t *testing.T) {
	_, err := InitDB("/nonexistent/dir/tasks.db")
	if err == nil {
		t.Fatal("Expected error for invalid path, got nil")
	}
}
