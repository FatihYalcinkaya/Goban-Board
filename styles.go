package main

import "github.com/charmbracelet/lipgloss"

var (
	columnStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	focusedStyle = columnStyle.Copy().
			BorderForeground(lipgloss.Color("205"))

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)
