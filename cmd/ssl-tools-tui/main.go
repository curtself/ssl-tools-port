// cmd/ssl-tools-tui/main.go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"ssl-tools/tui"
)

func main() {
	p := tea.NewProgram(tui.NewModel())
	if _,err := p.Run(); err != nil {
		fmt.Println("Error running program", err)
		os.Exit(1)
	}
}
