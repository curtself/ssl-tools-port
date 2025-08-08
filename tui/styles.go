// tui/styles.go
package tui
import "github.com/charmbracelet/lipgloss"

var (
	FocusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	BlurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	CursorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	ButtonFocusedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("57")).Padding(0, 2).Bold(true)
	ButtonBlurredStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 2)
	LabelStyle          = lipgloss.NewStyle().Bold(true)
	FieldWidth          = 40
)
