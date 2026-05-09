package main

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	defaultState state = iota
	inputState
	confirmState
	helpState
)

type confirmAction int

const (
	confirmNone confirmAction = iota
	confirmDeleteTask
	confirmDeleteColumn
)

type undoItem struct {
	task   Task
	column int
}

const maxUndoBuffer = 20

type RootModel struct {
	db            *sql.DB
	columns       []Column
	focusedColumn int
	state         state
	input         textinput.Model
	quitting      bool
	width         int
	height        int
	errMsg        string

	oldTitle        string
	editingTaskID   int
	editingDesc     bool
	columnRenameIdx int

	confirmAction confirmAction
	confirmTaskID int
	confirmColIdx int

	undoBuffer []undoItem
}

func (m RootModel) Init() tea.Cmd {
	return nil
}

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

		if m.state == inputState {
			switch msg.String() {
			case "enter":
				if m.input.Value() != "" {
					newVal := m.input.Value()
					switch {
					case m.oldTitle != "":
						items := m.columns[m.focusedColumn].list.Items()
						idx := m.columns[m.focusedColumn].list.Index()
						desc := ""
						if idx < len(items) {
							if t, ok := items[idx].(Task); ok {
								desc = t.description
							}
						}
						if err := RenameTask(m.db, m.editingTaskID, newVal); err != nil {
							m.errMsg = "Failed to rename task: " + err.Error()
						} else {
							items[idx] = NewTask(m.editingTaskID, newVal, desc)
							m.columns[m.focusedColumn].list.SetItems(items)
						}
					case m.editingDesc:
						if err := UpdateTaskDescription(m.db, m.editingTaskID, newVal); err != nil {
							m.errMsg = "Failed to update description: " + err.Error()
						} else {
							items := m.columns[m.focusedColumn].list.Items()
							for i, item := range items {
								t, ok := item.(Task)
								if ok && t.id == m.editingTaskID {
									items[i] = NewTask(t.id, t.title, newVal)
									break
								}
							}
							m.columns[m.focusedColumn].list.SetItems(items)
						}
						m.editingTaskID = 0
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
							m.columns[m.focusedColumn].list.InsertItem(0, NewTask(int(id), newVal, ""))
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

		m.errMsg = ""

		if len(m.columns) == 0 {
			switch msg.String() {
			case "n", "A":
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

		if m.columns[m.focusedColumn].list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.columns[m.focusedColumn].list, cmd = m.columns[m.focusedColumn].list.Update(msg)
			return m, cmd
		}

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

		case "n", "A":
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
			if index < len(curCol.list.Items())-1 {
				selectedItem := curCol.list.SelectedItem()
				curCol.list.RemoveItem(index)
				curCol.list.InsertItem(index+1, selectedItem)
				curCol.list.Select(index + 1)
			}

		case "ctrl+k":
			curCol := &m.columns[m.focusedColumn]
			index := curCol.list.Index()
			if index > 0 {
				selectedItem := curCol.list.SelectedItem()
				curCol.list.RemoveItem(index)
				curCol.list.InsertItem(index-1, selectedItem)
				curCol.list.Select(index - 1)
			}

		case "e":
			if selectedItem := m.columns[m.focusedColumn].list.SelectedItem(); selectedItem != nil {
				task, ok := selectedItem.(Task)
				if !ok {
					return m, nil
				}
				m.editingTaskID = task.id
				m.editingDesc = true
				m.input.Placeholder = "Enter description..."
				m.input.SetValue(task.description)
				m.state = inputState
				m.input.Focus()
			}

		case "u":
			if len(m.undoBuffer) > 0 {
				last := m.undoBuffer[len(m.undoBuffer)-1]
				m.undoBuffer = m.undoBuffer[:len(m.undoBuffer)-1]
				newID, err := UndoDeleteTask(m.db, last.task, last.column)
				if err != nil {
					m.errMsg = "Failed to undo delete: " + err.Error()
				} else if last.column >= 0 && last.column < len(m.columns) {
					m.columns[last.column].list.InsertItem(0, NewTask(int(newID), last.task.title, last.task.description))
				}
			}

		case "?":
			m.state = helpState
			return m, nil
		}
	}

	var cmd tea.Cmd
	if len(m.columns) > 0 {
		m.columns[m.focusedColumn].list, cmd = m.columns[m.focusedColumn].list.Update(msg)
	}
	return m, cmd
}

func (m *RootModel) resetState() RootModel {
	m.input.Blur()
	m.input.Reset()
	m.state = defaultState
	m.oldTitle = ""
	m.editingTaskID = 0
	m.editingDesc = false
	m.columnRenameIdx = -1
	return *m
}

func (m *RootModel) syncDimensions() {
	numCols := len(m.columns)
	if numCols == 0 {
		return
	}

	dynWidth := (m.width / numCols) - 4
	if dynWidth < 20 {
		dynWidth = 20
	}

	for i := range m.columns {
		m.columns[i].list.SetSize(dynWidth, m.height-12)

		m.columns[i].list.Styles.Title = m.columns[i].list.Styles.Title.
			Width(dynWidth).
			MaxWidth(dynWidth)

		m.columns[i].list.Styles.NoItems = m.columns[i].list.Styles.NoItems.
			Width(dynWidth).
			MaxWidth(dynWidth)
	}
}

func (m RootModel) View() string {
	if m.quitting {
		return ""
	}

	numCols := len(m.columns)
	if numCols == 0 {
		return "No columns"
	}

	surfaceStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Background(appBgHex).
		Align(lipgloss.Left, lipgloss.Top)

	if m.state == helpState {
		return surfaceStyle.Render(m.helpView())
	}

	minRequiredWidth := numCols * 25
	if m.width < minRequiredWidth || m.height < 15 {
		errorContent := lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.NewStyle().Foreground(pinkHex).Bold(true).Background(appBgHex).Render("TERMINAL TOO SMALL"),
			lipgloss.NewStyle().Foreground(grayHex).Background(appBgHex).Render("Please enlarge the window"),
		)
		return surfaceStyle.Align(lipgloss.Center, lipgloss.Center).Render(errorContent)
	}

	var views []string
	dynWidth := (m.width / numCols) - 2

	for i := range m.columns {
		m.columns[i].delegate.Styles.NormalTitle = normalTitleStyle
		m.columns[i].delegate.Styles.NormalDesc = normalDescStyle

		if i == m.focusedColumn {
			m.columns[i].delegate.Styles.SelectedTitle = focusedTitleStyle
			m.columns[i].delegate.Styles.SelectedDesc = focusedDescStyle
		} else {
			m.columns[i].delegate.Styles.SelectedTitle = selectedTitleStyle
			m.columns[i].delegate.Styles.SelectedDesc = selectedDescStyle
		}
		m.columns[i].list.SetDelegate(m.columns[i].delegate)

		m.columns[i].list.Styles.Title = m.columns[i].list.Styles.Title.Width(dynWidth).MaxWidth(dynWidth)

		style := columnStyle.Width(dynWidth)
		if i == m.focusedColumn {
			style = focusedStyle.Width(dynWidth)
		}
		views = append(views, style.Render(m.columns[i].list.View()))
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, views...)

	footerBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purpleHex).
		Padding(0, 1).
		Background(appBgHex)

	footerContainerStyle := lipgloss.NewStyle().
		Background(appBgHex).
		Width(m.width).
		Align(lipgloss.Center).
		MarginTop(1)

	var footerContent string
	if m.state == confirmState {
		switch m.confirmAction {
		case confirmDeleteTask:
			footerContent = " Delete task? (y/n) "
		case confirmDeleteColumn:
			colName := m.columns[m.focusedColumn].list.Title
			footerContent = fmt.Sprintf(" Delete column '%s' and all its tasks? (y/n) ", colName)
		}
	} else if m.errMsg != "" {
		footerContent = lipgloss.NewStyle().Foreground(pinkHex).Render("ERROR: " + m.errMsg)
	} else if m.state == inputState {
		footerContent = " " + m.input.View()
	} else {
		footerContent = "h/l: nav | j/k: move | ctrl+h/l/j/k: transfer | a: add | A: col-add | r: rename | e: desc | d: delete | R: col-rename | D: col-del | u: undo | ?: help | q: quit"

		contentMaxWidth := m.width - 6
		if contentMaxWidth < 20 {
			contentMaxWidth = 20
		}
		if len(footerContent) > contentMaxWidth {
			var lines []string
			for i := 0; i < len(footerContent); i += contentMaxWidth {
				end := i + contentMaxWidth
				if end > len(footerContent) {
					end = len(footerContent)
				}
				lines = append(lines, footerContent[i:end])
			}
			footerContent = strings.Join(lines, "\n")
		}
	}

	footer := footerContainerStyle.Render(footerBoxStyle.Render(footerContent))

	fullUI := lipgloss.JoinVertical(lipgloss.Left, board, footer)

	return surfaceStyle.Render(fullUI)
}

func (m RootModel) helpView() string {
	help := `Keyboard Shortcuts:

  Navigation:
    h/l              Navigate columns left/right
    j/k              Navigate tasks within a column

  Tasks:
    a                Add new task
    r                Rename selected task
    e                Edit description of selected task
    d                Delete selected task (with confirmation)
    ctrl+h/l         Move task to left/right column
    ctrl+j/k         Reorder task up/down within column

  Columns:
    n / A            Add new column
    R                Rename focused column
    D                Delete focused column (with confirmation)

  Other:
    /                Search/filter tasks in focused column
    u                Undo last deletion
    ?                Toggle this help
    q / ctrl+c       Quit

  Press ? or esc to close this help.`

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(purpleHex).
		Padding(1, 2).
		Background(appBgHex).
		Foreground(whiteHex)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Background(appBgHex).
		Align(lipgloss.Center, lipgloss.Center).
		Render(box.Render(help))
}
