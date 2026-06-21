package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.syncDimensions()

	case tea.KeyMsg:
		if m.state == helpState {
			switch msg.String() {
			case "?", "esc", "q", "ctrl+c":
				m.state = defaultState
			}
			return m, nil
		}

		if m.state == confirmState {
			return m.handleConfirmKey(msg)
		}

		if m.state == inputState {
			return m.handleInputKey(msg)
		}

		m.errMsg = ""

		if len(m.columns) == 0 {
			return m.handleNoColumnsKey(msg)
		}

		if m.columns[m.focusedColumn].list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.columns[m.focusedColumn].list, cmd = m.columns[m.focusedColumn].list.Update(msg)
			return m, cmd
		}

		return m.handleDefaultKey(msg)
	}

	var cmd tea.Cmd
	if len(m.columns) > 0 {
		m.columns[m.focusedColumn].list, cmd = m.columns[m.focusedColumn].list.Update(msg)
	}
	return m, cmd
}

func (m *RootModel) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		switch m.confirmAction {
		case confirmDeleteTask:
			for _, item := range m.columns[m.focusedColumn].list.Items() {
				t, ok := item.(Task)
				if ok && t.id == m.confirmTaskID {
					m.undoBuffer = append(m.undoBuffer, undoItem{task: t, column: m.focusedColumn})
					if len(m.undoBuffer) > maxUndoBuffer {
						m.undoBuffer = m.undoBuffer[1:]
					}
					break
				}
			}
			if err := DeleteTask(m.db, m.confirmTaskID); err != nil {
				m.errMsg = "Failed to delete task: " + err.Error()
			} else {
				index := m.columns[m.focusedColumn].list.Index()
				m.columns[m.focusedColumn].list.RemoveItem(index)
			}
		case confirmDeleteColumn:
			colIdx := m.confirmColIdx
			if err := DeleteTasksByStatus(m.db, colIdx); err != nil {
				m.errMsg = "Failed to delete column tasks: " + err.Error()
			} else if err := ShiftTaskStatuses(m.db, colIdx); err != nil {
				m.errMsg = "Failed to reindex remaining tasks: " + err.Error()
			} else if err := DeleteColumnByPosition(m.db, colIdx); err != nil {
				m.errMsg = "Failed to delete column: " + err.Error()
			} else if err := ShiftColumnPositions(m.db, colIdx); err != nil {
				m.errMsg = "Failed to reindex columns: " + err.Error()
			}
			if colIdx >= 0 && colIdx < len(m.columns) {
				m.columns = append(m.columns[:colIdx], m.columns[colIdx+1:]...)
				if len(m.columns) == 0 {
					m.focusedColumn = 0
				} else if m.focusedColumn >= len(m.columns) {
					m.focusedColumn = len(m.columns) - 1
				}
				m.syncDimensions()
			}
		}
		m.confirmAction = confirmNone
		m.state = defaultState
	case "n", "N", "esc":
		m.confirmAction = confirmNone
		m.state = defaultState
	}
	return m, nil
}

func (m *RootModel) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.input.Value() != "" {
			newVal := m.input.Value()
			switch {
			case m.oldTitle != "":
				items := m.columns[m.focusedColumn].list.Items()
				idx := m.columns[m.focusedColumn].list.Index()
				if err := RenameTask(m.db, m.editingTaskID, newVal); err != nil {
					m.errMsg = "Failed to rename task: " + err.Error()
				} else {
					items[idx] = NewTask(m.editingTaskID, newVal)
					m.columns[m.focusedColumn].list.SetItems(items)
				}
			case m.columnRenameIdx >= 0:
				if err := RenameColumn(m.db, m.columnRenameIdx, newVal); err != nil {
					m.errMsg = "Failed to rename column: " + err.Error()
				} else {
					m.columns[m.columnRenameIdx].list.Title = newVal
				}
				m.columnRenameIdx = -1
			default:
				id, err := SaveTask(m.db, newVal, m.focusedColumn)
				if err != nil {
					m.errMsg = "Failed to save task: " + err.Error()
				} else {
					m.columns[m.focusedColumn].list.InsertItem(0, NewTask(int(id), newVal))
				}
			}
		}
		return m.resetState(), nil
	case "esc":
		return m.resetState(), nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *RootModel) handleNoColumnsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "A":
		m.columns = append(m.columns, NewColumn("New Column"))
		if err := SaveColumn(m.db, "New Column", 0); err != nil {
			m.errMsg = "Failed to save column: " + err.Error()
		}
		m.focusedColumn = 0
		m.syncDimensions()
		return m, nil
	case "?":
		m.state = helpState
		return m, nil
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m *RootModel) handleDefaultKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit

	case "left", "h":
		if m.focusedColumn > 0 {
			m.focusedColumn--
		}

	case "right", "l":
		if m.focusedColumn < len(m.columns)-1 {
			m.focusedColumn++
		}

	case "H":
		if m.focusedColumn > 0 {
			other := m.focusedColumn - 1
			if err := SwapColumnPositions(m.db, m.focusedColumn, other); err != nil {
				m.errMsg = "Failed to swap column positions: " + err.Error()
			} else if err := SwapTaskStatuses(m.db, m.focusedColumn, other); err != nil {
				m.errMsg = "Failed to swap task statuses: " + err.Error()
			} else {
				m.columns[m.focusedColumn], m.columns[other] = m.columns[other], m.columns[m.focusedColumn]
				m.focusedColumn = other
				m.syncDimensions()
			}
		}

	case "L":
		if m.focusedColumn < len(m.columns)-1 {
			other := m.focusedColumn + 1
			if err := SwapColumnPositions(m.db, m.focusedColumn, other); err != nil {
				m.errMsg = "Failed to swap column positions: " + err.Error()
			} else if err := SwapTaskStatuses(m.db, m.focusedColumn, other); err != nil {
				m.errMsg = "Failed to swap task statuses: " + err.Error()
			} else {
				m.columns[m.focusedColumn], m.columns[other] = m.columns[other], m.columns[m.focusedColumn]
				m.focusedColumn = other
				m.syncDimensions()
			}
		}

	case "a":
		m.state = inputState
		m.input.Placeholder = "Enter task title..."
		m.input.Focus()
		return m, nil

	case "r":
		if len(m.columns[m.focusedColumn].list.Items()) > 0 {
			selected, ok := m.columns[m.focusedColumn].list.SelectedItem().(Task)
			if !ok {
				return m, nil
			}
			m.editingTaskID = selected.id
			m.oldTitle = selected.title
			m.input.Placeholder = "Rename task..."
			m.input.SetValue(selected.title)
			m.state = inputState
			m.input.Focus()
			return m, nil
		}

	case "R":
		m.columnRenameIdx = m.focusedColumn
		m.input.Placeholder = "Enter column name..."
		m.input.SetValue(m.columns[m.focusedColumn].list.Title)
		m.state = inputState
		m.input.Focus()
		return m, nil

	case "d":
		if selectedItem := m.columns[m.focusedColumn].list.SelectedItem(); selectedItem != nil {
			task, ok := selectedItem.(Task)
			if !ok {
				return m, nil
			}
			m.state = confirmState
			m.confirmAction = confirmDeleteTask
			m.confirmTaskID = task.id
		}

	case "D":
		m.state = confirmState
		m.confirmAction = confirmDeleteColumn
		m.confirmColIdx = m.focusedColumn

	case "A":
		m.columns = append(m.columns, NewColumn(fmt.Sprintf("New Column %d", len(m.columns)+1)))
		if err := SaveColumn(m.db, m.columns[len(m.columns)-1].list.Title, len(m.columns)-1); err != nil {
			m.errMsg = "Failed to save column: " + err.Error()
		}
		m.syncDimensions()

	case "ctrl+l":
		if m.focusedColumn < len(m.columns)-1 {
			selectedItem := m.columns[m.focusedColumn].list.SelectedItem()
			if selectedItem != nil {
				task, ok := selectedItem.(Task)
				if !ok {
					return m, nil
				}
				if err := UpdateTaskStatus(m.db, task.id, m.focusedColumn+1); err != nil {
					m.errMsg = "Failed to move task: " + err.Error()
				} else {
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn+1].list.InsertItem(0, selectedItem)
					m.focusedColumn++
				}
			}
		}

	case "ctrl+h":
		if m.focusedColumn > 0 {
			selectedItem := m.columns[m.focusedColumn].list.SelectedItem()
			if selectedItem != nil {
				task, ok := selectedItem.(Task)
				if !ok {
					return m, nil
				}
				if err := UpdateTaskStatus(m.db, task.id, m.focusedColumn-1); err != nil {
					m.errMsg = "Failed to move task: " + err.Error()
				} else {
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn-1].list.InsertItem(0, selectedItem)
					m.focusedColumn--
				}
			}
		}

	case "ctrl+j":
		curCol := &m.columns[m.focusedColumn]
		index := curCol.list.Index()
		items := curCol.list.Items()
		if index < len(items)-1 {
			taskA := items[index].(Task)
			taskB := items[index+1].(Task)
			selectedItem := curCol.list.SelectedItem()
			curCol.list.RemoveItem(index)
			curCol.list.InsertItem(index+1, selectedItem)
			curCol.list.Select(index + 1)
			if err := SwapTaskPositions(m.db, taskA.id, taskB.id); err != nil {
				m.errMsg = "Failed to persist position: " + err.Error()
			}
		}

	case "ctrl+k":
		curCol := &m.columns[m.focusedColumn]
		index := curCol.list.Index()
		items := curCol.list.Items()
		if index > 0 {
			taskA := items[index].(Task)
			taskB := items[index-1].(Task)
			selectedItem := curCol.list.SelectedItem()
			curCol.list.RemoveItem(index)
			curCol.list.InsertItem(index-1, selectedItem)
			curCol.list.Select(index - 1)
			if err := SwapTaskPositions(m.db, taskA.id, taskB.id); err != nil {
				m.errMsg = "Failed to persist position: " + err.Error()
			}
		}

	case "u":
		if len(m.undoBuffer) > 0 {
			last := m.undoBuffer[len(m.undoBuffer)-1]
			m.undoBuffer = m.undoBuffer[:len(m.undoBuffer)-1]
			newID, err := UndoDeleteTask(m.db, last.task, last.column)
			if err != nil {
				m.errMsg = "Failed to undo delete: " + err.Error()
			} else if last.column >= 0 && last.column < len(m.columns) {
				m.columns[last.column].list.InsertItem(0, NewTask(int(newID), last.task.title))
			}
		}

	case "?":
		m.state = helpState
		return m, nil
	}

	return m, nil
}
