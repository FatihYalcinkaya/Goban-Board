package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type Column struct {
	list list.Model
}

func NewColumn(title string) Column {
	d := list.NewDefaultDelegate()

	// --- TEMEL STİL (VARSAYILAN BEYAZ) ---
	// Başlangıçta tüm itemlar (seçili olsa bile) beyaz görünsün.
	// Renk değişimini model.go'nun View fonksiyonunda dinamik yapacağız.
	whiteStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		PaddingLeft(0).
		MarginLeft(0).
		BorderLeft(false)

	d.Styles.NormalTitle = whiteStyle
	d.Styles.SelectedTitle = whiteStyle // Seçili olsa da ilk hali beyaz

	d.Styles.NormalDesc = whiteStyle.Foreground(lipgloss.Color("245"))
	d.Styles.SelectedDesc = whiteStyle.Foreground(lipgloss.Color("245"))

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = title

	// --- ÜST BAŞLIK (ORTALANMIŞ VE SIFIR MARGIN) ---
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true).
		Align(lipgloss.Center).
		Padding(0, 0).
		Margin(0, 0)

	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return Column{list: l}
}
