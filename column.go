package main

import "github.com/charmbracelet/bubbles/list"

type Column struct {
	list list.Model
}

func NewColumn(title string) Column {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 20, 10)
	l.Title = title
	l.SetShowHelp(false)
	return Column{list: l}
}
