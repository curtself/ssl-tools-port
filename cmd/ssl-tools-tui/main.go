// cmd/ssl-tools-tui/main.go
package ssl_tools_tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"ssl-tools/tui"
)

func RunTUI() {
	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())
	if _,err := p.Run(); err != nil {
		fmt.Println("Error running program", err)
		os.Exit(1)
	}
}
