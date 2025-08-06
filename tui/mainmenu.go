// tui/mainmenu.go
package tui

import (
	"github.com/charmbracelet/bubbles/list"
)

type menuItem string

func (i menuItem) Title() string       { return string(i) }
func (i menuItem) Description() string { return "" }
func (i menuItem) FilterValue() string { return string(i) }

func NewMainMenu(width, height int) list.Model {
	items := []list.Item{
		menuItem("Create"),
		menuItem("Finish"),
		menuItem("Info"),
		menuItem("Exit"),
	}
	l := list.New(items, list.NewDefaultDelegate(), width, height) // don't have width or height....
	l.Title = "SSL Tools - Select Operation"
	l.SetFilteringEnabled(false)
	return l
}
