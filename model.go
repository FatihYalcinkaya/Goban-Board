package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	defaultState state = iota
	inputState
)

type RootModel struct {
	columns       []Column
	focusedColumn int
	state         state
	input         textinput.Model
	quitting      bool
	width         int
	height        int
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
		// --- Input Mode Logic ---
		if m.state == inputState {
			switch msg.String() {
			case "enter":
				if m.input.Value() != "" {
					m.columns[m.focusedColumn].list.InsertItem(0, NewTask(m.input.Value(), ""))
				}
				return m.resetState(), nil
			case "esc":
				return m.resetState(), nil
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}

		// --- Normal Mode Logic ---
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		// Column Navigation
		case "left", "h":
			if m.focusedColumn > 0 {
				m.focusedColumn--
			}
		case "right", "l":
			if m.focusedColumn < len(m.columns)-1 {
				m.focusedColumn++
			}

		// Task Actions
		case "a":
			m.state = inputState
			m.input.Focus()
			return m, nil

		case "r":
			if len(m.columns[m.focusedColumn].list.Items()) > 0 {
				selected := m.columns[m.focusedColumn].list.SelectedItem().(Task)
				m.input.SetValue(selected.title)
				m.state = inputState
				m.input.Focus()
				m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
				return m, nil
			}

		case "d":
			if len(m.columns[m.focusedColumn].list.Items()) > 0 {
				index := m.columns[m.focusedColumn].list.Index()
				m.columns[m.focusedColumn].list.RemoveItem(index)
			}

		case "n":
			m.columns = append(m.columns, NewColumn("New Column"))
			m.syncDimensions()

		// Transfer Logic (Sütunlar Arası)
		case "ctrl+l":
			if m.focusedColumn < len(m.columns)-1 {
				selectedItem := m.columns[m.focusedColumn].list.SelectedItem()
				if selectedItem != nil {
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn+1].list.InsertItem(0, selectedItem)
					m.focusedColumn++
				}
			}

		case "ctrl+h":
			if m.focusedColumn > 0 {
				selectedItem := m.columns[m.focusedColumn].list.SelectedItem()
				if selectedItem != nil {
					m.columns[m.focusedColumn].list.RemoveItem(m.columns[m.focusedColumn].list.Index())
					m.columns[m.focusedColumn-1].list.InsertItem(0, selectedItem)
					m.focusedColumn--
				}
			}

		// Liste İçi Taşıma (Dikey)
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
	return *m
}

func (m *RootModel) syncDimensions() {
	numCols := len(m.columns)
	if numCols == 0 {
		return
	}

	dynWidth := (m.width / numCols) - 5
	if dynWidth < 20 {
		dynWidth = 20
	}

	for i := range m.columns {
		m.columns[i].list.SetSize(dynWidth, m.height-12)
		// Başlık genişliği için ince ayar
		m.columns[i].list.Styles.Title = m.columns[i].list.Styles.Title.
			Width(dynWidth - 4).
			MaxWidth(dynWidth - 4)
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

	// --- EKRAN BOYUTU KONTROLÜ ---
	// Her sütun için min 25 karakter genişlik ve 15 karakter yükseklik sınırı
	minRequiredWidth := numCols * 25
	minRequiredHeight := 15

	if m.width < minRequiredWidth || m.height < minRequiredHeight {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("201")).
			Bold(true)

		subStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

		content := lipgloss.JoinVertical(
			lipgloss.Center,
			errorStyle.Render("TERMINAL TOO SMALL"),
			subStyle.Render("Please enlarge to view the board"),
		)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	// --- RENDER MANTIĞI ---
	var views []string
	dynWidth := (m.width / numCols) - 5
	if dynWidth < 20 {
		dynWidth = 20
	}

	for i := range m.columns {
		// --- TAM MİNİMAL DELEGAT AYARI ---
		d := list.NewDefaultDelegate()

		// Ortak Ayarlar: Çizgiyi ve kaymayı her durumda kapat
		d.Styles.NormalTitle = d.Styles.NormalTitle.PaddingLeft(0).MarginLeft(0).Foreground(lipgloss.Color("255"))
		d.Styles.NormalDesc = d.Styles.NormalDesc.PaddingLeft(0).MarginLeft(0).Foreground(lipgloss.Color("245"))

		// Seçili stilin dikey çizgisini ve padding'ini sıfırlıyoruz
		d.Styles.SelectedTitle = d.Styles.SelectedTitle.
			BorderLeft(false).
			PaddingLeft(0).
			MarginLeft(0)
		d.Styles.SelectedDesc = d.Styles.SelectedDesc.
			BorderLeft(false).
			PaddingLeft(0).
			MarginLeft(0)

		// Renk Ayarı: Sadece odaklanılan sütun renkli (pembe/mor) olsun
		if i == m.focusedColumn {
			d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(lipgloss.Color("205")).Bold(true)
			d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(lipgloss.Color("205"))
		} else {
			d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(lipgloss.Color("255")).Bold(false)
			d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(lipgloss.Color("245"))
		}

		m.columns[i].list.SetDelegate(d)

		// Sütun Stili (Deprecated Copy kaldırıldı)
		style := columnStyle.Width(dynWidth)
		if i == m.focusedColumn {
			style = focusedStyle.Width(dynWidth)
		}
		views = append(views, style.Render(m.columns[i].list.View()))
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, views...)

	// Footer (İngilizce)
	var footer string
	if m.state == inputState {
		footer = "\n Edit/Add: " + m.input.View()
	} else {
		footer = helpStyle.Render("\n h/l/j/k: move | ctrl+h/l/j/k: transfer | a: add | r: rename | d: delete | q: quit")
	}

	fullUI := lipgloss.JoinVertical(lipgloss.Center, board, footer)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, fullUI)
}
