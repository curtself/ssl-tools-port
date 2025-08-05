// tui/model.go
package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

// BackToMenuMsg signals going back to the main menu.
type BackToMenuMsg struct{}

const (
	stateMainMenu state = iota
	stateCreate
	stateFinish
	stateInfo
	stateExit
)

type Model struct {
	state       state
	width       int
	height      int
	mainMenu    list.Model
	createModel *CreateModel
	finishModel *FinishModel
	infoModel   *InfoModel
	menuInit	bool
}

func NewModel() Model {
	return Model{
		state: stateMainMenu,
		//mainMenu: NewMainMenu(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateMainMenu:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			if !m.menuInit {
				m.mainMenu = NewMainMenu(m.width,m.height)
				m.menuInit = true
			}
			//m.mainMenu.SetSize(msg.Width, msg.Height)
			return m, nil
		}
		var cmd tea.Cmd
		m.mainMenu, cmd = m.mainMenu.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			key := keyMsg.String()
			switch key {
			case "enter":
				if selected, ok := m.mainMenu.SelectedItem().(menuItem); ok {
					switch selected {
					case "Create":
						m.state = stateCreate
						m.createModel = NewCreateModel()
					case "Finish":
						m.state = stateFinish
						m.finishModel = NewFinishModel()
					case "Info":
						m.state = stateInfo
						m.infoModel = NewInfoModel()
					case "Exit":
						m.state = stateExit
						return m, tea.Quit
					}
				}
			case "c", "C":
				m.state = stateCreate
				m.createModel = NewCreateModel()
			case "f", "F":
				m.state = stateFinish
				m.finishModel = NewFinishModel()
			case "i", "I":
				m.state = stateInfo
				m.infoModel = NewInfoModel()
			case "e", "E":
				m.state = stateExit
				return m, tea.Quit
			}
		}
		return m, cmd
	case stateCreate:
		updatedModel, cmd := m.createModel.Update(msg)
		m.createModel = updatedModel.(*CreateModel)
		switch msg.(type) {
		case BackToMenuMsg:
			m.state = stateMainMenu
			m.createModel = nil // cleans up resources
			return m, nil
		}
		//return m.createModel.Update(msg)
		return m, cmd
	case stateFinish:
		updatedModel, cmd := m.finishModel.Update(msg)
		m.finishModel = updatedModel.(*FinishModel)
		switch msg.(type) {
		case BackToMenuMsg:
			m.state = stateMainMenu
			m.finishModel = nil
			return m, nil
		}
		//return m.finishModel.Update(msg)
		return m, cmd
	case stateInfo:
		updatedModel, cmd := m.infoModel.Update(msg)
		m.infoModel = updatedModel.(*InfoModel)
		switch msg.(type) {
		case BackToMenuMsg:
			m.state = stateMainMenu
			m.infoModel = nil
			return m, nil
		}
		//return m.infoModel.Update(msg)
		return m, cmd
	default:
		return m, tea.Quit
	}
}

func (m Model) View() string {
	switch m.state {
	case stateMainMenu:
		if m.menuInit {
			return m.mainMenu.View()
		} else {
			return ""
		}
	case stateCreate:
		return m.createModel.View()
	case stateFinish:
		return m.finishModel.View()
	case stateInfo:
		return m.infoModel.View()
	default:
		return "Exiting...\n"
	}
}
