package main

import "github.com/charmbracelet/bubbles/list"

type Task struct {
	title       string
	description string
}

func (t Task) Title() string       { return t.title }
func (t Task) Description() string { return t.description }
func (t Task) FilterValue() string { return t.title }

func NewTask(title, desc string) list.Item {
	return Task{title: title, description: desc}
}
