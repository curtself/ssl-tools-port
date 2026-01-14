// tui/styles.go
package tui
import "github.com/charmbracelet/lipgloss"

var (
	//FocusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	FocusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	//BlurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	BlurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	//CursorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	CursorStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("135"))
	MsgSuccessStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	MsgErrorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	ButtonFocusedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("57")).Padding(0, 2).Bold(true)
	ButtonBlurredStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 2)
	LabelStyle          = lipgloss.NewStyle().Bold(true)
	FieldWidth          = 40
)
