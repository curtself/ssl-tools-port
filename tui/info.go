// tui/info.go
package tui

import tea "github.com/charmbracelet/bubbletea"

type InfoModel struct {}

func NewInfoModel() *InfoModel {
	return &InfoModel{}
}

func (m *InfoModel) Init() tea.Cmd {
	return nil
}

func (m *InfoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m.handleBack()
		}
	}
	return m, nil
}

func (m *InfoModel) handleBack() (tea.Model, tea.Cmd) {
	// TODO: return a custom Msg so parent Model can switch back to main menu
	// For now just quit to demonstrate:
	//return m, tea.Quit
	return m, func() tea.Msg {
		return BackToMenuMsg{}
	}
}

func (m *InfoModel) View() string {
	return "[Certificate Info] Coming soon...\n"
}

