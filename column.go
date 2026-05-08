package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type Column struct {
	list     list.Model
	delegate list.DefaultDelegate
}

func NewColumn(title string) Column {
	d := list.NewDefaultDelegate()

	itemStyle := lipgloss.NewStyle().
		Background(appBgHex).
		Foreground(whiteHex)

	d.Styles.NormalTitle = itemStyle
	d.Styles.SelectedTitle = itemStyle.Copy().Bold(true)
	d.Styles.NormalDesc = itemStyle.Copy().Foreground(grayHex)
	d.Styles.SelectedDesc = itemStyle.Copy().Foreground(grayHex)

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = title

	l.Styles.Title = lipgloss.NewStyle().
		Background(purpleHex).
		Foreground(whiteLiteral).
		Bold(true).
		Padding(0, 1)

	l.Styles.NoItems = lipgloss.NewStyle().
		Background(appBgHex).
		Foreground(grayHex)

	l.Styles.HelpStyle = lipgloss.NewStyle().
		Background(appBgHex)

	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return Column{list: l, delegate: d}
}
