// tui/create.go
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"ssl-tools/internal/options"
	"ssl-tools/internal/certsvc"
)

var createOpts options.CreateOptions

var keySizes = []int{2048, 3072, 4096}

type CreateModel struct {
	inputs       []textinput.Model
	keySizeIndex int
	focused      int  // current focus index over all fields + buttons
	editing      bool // true if currently editing a text input
	statusMsg	string
	success		bool
}

const (
	fieldCommonName 	= 0
	fieldSANs       	= 1
	fieldExternalKey 	= 2
	fieldKeySize		= 3
	buttonCreateCSR 	= 4
	buttonBack      	= 5
)

func NewCreateModel() *CreateModel {
	inputs := make([]textinput.Model, 3)

	// Common Name
	ti := textinput.New()
	ti.Placeholder = "Common Name (required)"
	ti.CharLimit = 256
	ti.Width = FieldWidth
	inputs[fieldCommonName] = ti

	// SAN list
	ti2 := textinput.New()
	ti2.Placeholder = "SANs (comma separated)"
	ti2.CharLimit = 512
	ti2.Width = FieldWidth
	inputs[fieldSANs] = ti2

	// External Key
	ti3 := textinput.New()
	ti3.Placeholder = "External Key File (optional)"
	ti3.CharLimit = 256
	ti3.Width = FieldWidth
	inputs[fieldExternalKey] = ti3

	return &CreateModel{
		inputs:       inputs,
		keySizeIndex: 2, // default to 4096
		focused:      0,
		editing:      false,
	}
}

func (m *CreateModel) Init() tea.Cmd {
	// Focus first input by default
	m.inputs[m.focused].Focus()
	return textinput.Blink
}

func (m *CreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q", "esc":
			return m.handleBack()

		case "up", "k":
			if m.editing {
				// pass key to textinput
				return m.updateInputs(msg)
			}
			m.focused--
			if m.focused < 0 {
				m.focused = buttonBack
			}
			m.updateFocus()
			return m, nil

		case "down", "j":
			if m.editing {
				return m.updateInputs(msg)
			}
			m.focused++
			if m.focused > buttonBack {
				m.focused = 0
			}
			m.updateFocus()
			return m, nil

		case "left", "h":
			if m.editing {
				return m.updateInputs(msg)
			}
			if m.focused == buttonCreateCSR || m.focused == buttonBack {
				return m, nil
			}
			if m.focused == fieldExternalKey || m.focused == fieldCommonName || m.focused == fieldSANs {
				// no left/right for text inputs, ignore
				return m, nil
			}
			if m.focused == fieldKeySize { // Key Size, cycle down
				m.keySizeIndex--
				if m.keySizeIndex < 0 {
					m.keySizeIndex = len(keySizes) - 1
				}
			}
			return m, nil

		case "right", "l":
			if m.editing {
				return m.updateInputs(msg)
			}
			if m.focused == buttonCreateCSR || m.focused == buttonBack {
				return m, nil
			}
			if m.focused == fieldExternalKey || m.focused == fieldCommonName || m.focused == fieldSANs {
				return m, nil
			}
			if m.focused == fieldKeySize { // Key Size, cycle up
				m.keySizeIndex++
				if m.keySizeIndex >= len(keySizes) {
					m.keySizeIndex = 0
				}
			}
			return m, nil

		case "enter":
			if m.editing {
				// finish editing
				m.editing = false
				m.inputs[m.focused].Blur()
				return m, nil
			}

			// Not editing, enter triggers edit mode or buttons
			switch m.focused {
			case fieldCommonName, fieldSANs, fieldExternalKey:
				m.editing = true
				m.inputs[m.focused].Focus()
				return m, textinput.Blink
			case fieldKeySize: // keySize - no editing, do nothing here
				return m, nil
			case buttonCreateCSR:
				// Run your CreateCSR logic here
				return m.handleCreate()
			case buttonBack:
				// Signal to parent to go back to main menu
				return m.handleBack()
			}
		}

	case tea.WindowSizeMsg:
		// ignore for now
	}

	if m.editing {
		return m.updateInputs(msg)
	}

	return m, nil
}

func (m *CreateModel) updateInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m *CreateModel) updateFocus() {
	for i := range m.inputs {
		if i == m.focused && !m.editing {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *CreateModel) View() string {
	var b strings.Builder

	b.WriteString("Use ↑/↓ (k/j) to navigate, Enter to edit/select, Left/Right to change Key Size\n\n")

	// Common Name
	b.WriteString(m.renderField("Common Name", m.inputs[fieldCommonName].Value(), fieldCommonName))

	// SANs
	b.WriteString(m.renderField("SAN list", m.inputs[fieldSANs].Value(), fieldSANs))

	// External Key
	b.WriteString(m.renderField("External Key", m.inputs[fieldExternalKey].Value(), fieldExternalKey))

	// Key Size
	b.WriteString(m.renderKeySize())

	// Buttons
	b.WriteString(m.renderButton("Create CSR", buttonCreateCSR))
	b.WriteString(m.renderButton("Back to Menu", buttonBack))

	if m.statusMsg != "" {
		if m.success {
			b.WriteString( lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("\n"+m.statusMsg+"\n"))
		} else {
			b.WriteString( lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("\n"+m.statusMsg+"\n"))
		}
	}
	return b.String()
}

func (m *CreateModel) renderField(label, value string, idx int) string {
	focused := m.focused == idx
	style := BlurredStyle
	if focused {
		style = FocusedStyle
	}

	// if editing, show the input.View()
	if m.editing && focused {
		return fmt.Sprintf("%s %s\n", LabelStyle.Render(label+":"),
			m.inputs[idx].View())
	}

	// else show label and value with cursor indicator if focused
	cursor := " "
	if focused {
		cursor = ">"
	}
	val := value
	if val == "" {
		val = "<none>"
	}
	return fmt.Sprintf("%s %s %s\n", CursorStyle.Render(cursor), style.Render(label+":"), style.Render(val))
}

func (m *CreateModel) renderKeySize() string {
	focused := m.focused == fieldKeySize
	style := BlurredStyle
	if focused {
		style = FocusedStyle
	}

	cursor := " "
	if focused {
		cursor = ">"
	}
	return fmt.Sprintf("%s %s %d\n",
		CursorStyle.Render(cursor),
		style.Render("Key Size:"),
		keySizes[m.keySizeIndex])
}

func (m *CreateModel) renderButton(label string, idx int) string {
	focused := m.focused == idx
	style := ButtonBlurredStyle
	if focused {
		style = ButtonFocusedStyle
	}

	return style.Render(fmt.Sprintf("[%s]", label))
}

// handleCreate triggers the CSR creation using the inputs and keysize
func (m *CreateModel) handleCreate() (tea.Model, tea.Cmd) {
	// Here you would construct your options.CreateOptions from the inputs and keySize
	// then call your certsvc service CreateCSR and SaveCSRdto like you described.
	createOpts.CommonName = m.inputs[fieldCommonName].Value()
	sanSlice := []string{}
	for _, s := range strings.Split(m.inputs[fieldSANs].Value()," ") {
		sanSlice = append( sanSlice, strings.ReplaceAll(s,",","") )
	}
	createOpts.SANs = sanSlice
	createOpts.KeySize = keySizes[m.keySizeIndex]
	createOpts.Key = m.inputs[fieldExternalKey].Value()
	// Need to call validate and return a custom message if it fails
	if err := createOpts.Validate(); err != nil {
		return m, func() tea.Msg {
			return ValidateErrorMsg{err:err}
		}
	}
	// instantiate the service and use it
	svc := certsvc.New()
	if err := svc.SetKeyLength(createOpts.KeySize); err != nil {
		return m, func() tea.Msg {
			return ValidateErrorMsg{err:err}
		}
	}
	result, err := svc.CreateCSR(createOpts)
	if err != nil {
		return m, func() tea.Msg {
			return CreateResultErrorMsg{err:err}
		}
	}
	outputLines, err := svc.SaveCSRdto(result)
	if err != nil {
		return m, func() tea.Msg {
			return CreateResultErrorMsg{err:err}
		}
	}
	m.success = true
	return m, func() tea.Msg {
		return SuccessMsg{logs: outputLines}
	}
}

// handleBack sends a Quit or a custom Msg to parent model to return to main menu
func (m *CreateModel) handleBack() (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		return BackToMenuMsg{}
	}
}

