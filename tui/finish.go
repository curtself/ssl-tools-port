// tui/finish.go
package tui

import tea "github.com/charmbracelet/bubbletea"

type FinishModel struct {}

func NewFinishModel() *FinishModel {
	return &FinishModel{}
}

func (m *FinishModel) Init() tea.Cmd {
	return nil
}

func (m *FinishModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m.handleBack()
		}
	}
	return m, nil
}

func (m *FinishModel) handleBack() (tea.Model, tea.Cmd) {
	// TODO: return a custom Msg so parent Model can switch back to main menu
	// For now just quit to demonstrate:
	//return m, tea.Quit
	return m, func() tea.Msg {
		return BackToMenuMsg{}
	}
}

func (m *FinishModel) View() string {
	return "[Finish Certificate] Coming soon...\n"
}

