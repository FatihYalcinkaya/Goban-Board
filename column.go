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

	// Satırların arka planını sabitle
	itemStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(appBgHex)).
		Foreground(lipgloss.Color(whiteHex))

	d.Styles.NormalTitle = itemStyle
	d.Styles.SelectedTitle = itemStyle.Copy().Bold(true)
	d.Styles.NormalDesc = itemStyle.Copy().Foreground(lipgloss.Color(grayHex))
	d.Styles.SelectedDesc = itemStyle.Copy().Foreground(lipgloss.Color(grayHex))

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = title

	// --- SIZINTI KESİCİ AYARLAR ---
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color(purpleHex)).
		Foreground(lipgloss.Color("#ffffff")).
		Bold(true).
		Padding(0, 1)

	// O üç noktaların ve "No items" yazısının etrafını boya
	l.Styles.NoItems = lipgloss.NewStyle().
		Background(lipgloss.Color(appBgHex)).
		Foreground(lipgloss.Color(grayHex))

	// Listenin altındaki görünmez yardım alanını da boya (o alttaki sızıntı için)
	l.Styles.HelpStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(appBgHex))

	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return Column{list: l}
}
