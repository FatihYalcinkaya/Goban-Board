package main

import "github.com/charmbracelet/lipgloss"

var (
	// Sütun genişliğini burada sabitliyoruz (örneğin 30 karakter)
	fixedColumnWidth = 30

	columnStyle = lipgloss.NewStyle().
			Width(fixedColumnWidth). // Sütun genişliğini zorla
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Align(lipgloss.Left) // Metni sola yasla

	focusedStyle = columnStyle.Copy().
			BorderForeground(lipgloss.Color("205"))

	// EKSİK OLAN DEĞİŞKEN BURASI:
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

