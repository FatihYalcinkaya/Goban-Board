package main

import "github.com/charmbracelet/lipgloss"

var (
	fixedColumnWidth = 30

	columnStyle = lipgloss.NewStyle().
			Width(fixedColumnWidth).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Align(lipgloss.Left)

	focusedStyle = columnStyle.Copy().
			BorderForeground(lipgloss.Color("205"))

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)
