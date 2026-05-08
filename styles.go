package main

import "github.com/charmbracelet/lipgloss"

var (
	// --- FIXED HEX COLORS ---
	pinkHex   = "#ff5faf"
	purpleHex = "#7d7aff"
	whiteHex  = "#eeeeee"
	grayHex   = "#8a8a8a"
	borderHex = "#5f5fdf"

	// Final Background Color
	appBgHex = "#1d1f21"

	fixedColumnWidth = 30

	// Default column style
	columnStyle = lipgloss.NewStyle().
			Width(fixedColumnWidth).
			Padding(1, 1).
			Background(lipgloss.Color(appBgHex)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(borderHex)).
		// This kills the terminal background behind the border lines
		BorderBackground(lipgloss.Color(appBgHex)).
		Align(lipgloss.Left)

	// Focused column style
	focusedStyle = lipgloss.NewStyle().
			Width(fixedColumnWidth).
			Padding(1, 1).
			Background(lipgloss.Color(appBgHex)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(pinkHex)).
			BorderBackground(lipgloss.Color(appBgHex)).
			Align(lipgloss.Left)

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(grayHex))
)
