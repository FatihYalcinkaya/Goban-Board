package main

import "github.com/charmbracelet/lipgloss"

var (
	// --- FIXED HEX COLORS (AdaptiveColor with same light/dark to preserve color scheme on any terminal) ---
	pinkHex   = lipgloss.AdaptiveColor{Light: "#ff5faf", Dark: "#ff5faf"}
	purpleHex = lipgloss.AdaptiveColor{Light: "#7d7aff", Dark: "#7d7aff"}
	whiteHex  = lipgloss.AdaptiveColor{Light: "#eeeeee", Dark: "#eeeeee"}
	grayHex   = lipgloss.AdaptiveColor{Light: "#8a8a8a", Dark: "#8a8a8a"}
	borderHex = lipgloss.AdaptiveColor{Light: "#5f5fdf", Dark: "#5f5fdf"}

	// Final Background Color
	appBgHex = lipgloss.AdaptiveColor{Light: "#1d1f21", Dark: "#1d1f21"}

	whiteLiteral = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"}

	fixedColumnWidth = 30

	// Default column style
	columnStyle = lipgloss.NewStyle().
			Width(fixedColumnWidth).
			Padding(1, 1).
			Background(appBgHex).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderHex).
		// This kills the terminal background behind the border lines
		BorderBackground(appBgHex).
		Align(lipgloss.Left)

	// Focused column style
	focusedStyle = lipgloss.NewStyle().
			Width(fixedColumnWidth).
			Padding(1, 1).
			Background(appBgHex).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(pinkHex).
			BorderBackground(appBgHex).
			Align(lipgloss.Left)

	helpStyle = lipgloss.NewStyle().Foreground(grayHex).Background(appBgHex)
)
