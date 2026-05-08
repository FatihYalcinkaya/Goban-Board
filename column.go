package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type Column struct {
	list list.Model
}

func NewColumn(title string) Column {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = title

	// Tüm varsayılan boşlukları (margin/padding) sıfırlıyoruz
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Bold(true).
		Align(lipgloss.Center).
		Padding(0, 0).
		Margin(0, 0) // BURASI KRİTİK: Basamağı yapan o margin'i sildik.

	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return Column{list: l}
}
