package main

import "github.com/charmbracelet/bubbles/list"

type Task struct {
	id    int
	title string
}

func (t Task) Title() string       { return t.title }
func (t Task) Description() string { return "" }
func (t Task) FilterValue() string { return t.title }

func NewTask(id int, title string) list.Item {
	return Task{id: id, title: title}
}
