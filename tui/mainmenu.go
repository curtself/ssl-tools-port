// tui/mainmenu.go
package tui

import (
	"github.com/charmbracelet/bubbles/list"
)

type menuItem struct {
	title	string
	descr	string
}

func (i menuItem) Title() string       { return string(i.title) }
func (i menuItem) Description() string { return string(i.descr) }
func (i menuItem) FilterValue() string { return string(i.title) }

func NewMainMenu(width, height int) list.Model {
	items := []list.Item{
		menuItem(menuItem{title:"Create",descr:"Create a CSR and key",}),
		menuItem(menuItem{title:"Finish",descr:"Combine an issued certificate and key",}),
		menuItem(menuItem{title:"Info",descr:"View certificate or CSR information",}),
		menuItem(menuItem{title:"Exit"}),
	}
	l := list.New(items, list.NewDefaultDelegate(), width, height) // don't have width or height....
	l.Title = "SSL Tools - Select Operation"
	l.SetFilteringEnabled(false)
	return l
}
