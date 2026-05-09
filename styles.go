package main

import "github.com/charmbracelet/lipgloss"

var (
	pinkHex   = lipgloss.AdaptiveColor{Light: "#ff5faf", Dark: "#ff5faf"}
	purpleHex = lipgloss.AdaptiveColor{Light: "#7d7aff", Dark: "#7d7aff"}
	whiteHex  = lipgloss.AdaptiveColor{Light: "#eeeeee", Dark: "#eeeeee"}
	grayHex   = lipgloss.AdaptiveColor{Light: "#8a8a8a", Dark: "#8a8a8a"}
	borderHex = lipgloss.AdaptiveColor{Light: "#5f5fdf", Dark: "#5f5fdf"}

	appBgHex = lipgloss.AdaptiveColor{Light: "#1d1f21", Dark: "#1d1f21"}

	whiteLiteral = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"}

	columnStyle = lipgloss.NewStyle().
			Width(30).
			Padding(1, 1).
			Background(appBgHex).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderHex).
			BorderBackground(appBgHex).
			Align(lipgloss.Left)

	focusedStyle = lipgloss.NewStyle().
			Width(30).
			Padding(1, 1).
			Background(appBgHex).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(pinkHex).
			BorderBackground(appBgHex).
			Align(lipgloss.Left)

	helpStyle = lipgloss.NewStyle().Foreground(grayHex).Background(appBgHex)

	normalTitleStyle   = lipgloss.NewStyle().Background(appBgHex).Foreground(whiteHex)
	selectedTitleStyle = lipgloss.NewStyle().Background(appBgHex).BorderLeft(false)
	focusedTitleStyle  = lipgloss.NewStyle().Background(appBgHex).BorderLeft(false).Foreground(pinkHex).Bold(true)
)
