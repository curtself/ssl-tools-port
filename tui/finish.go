package tui

import (
	//"fmt"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	//"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"ssl-tools/internal/certsvc"
	"ssl-tools/internal/options"
)

type finishFocus int

const (
	focusFinishCert finishFocus = iota
	focusFinishKey
	focusFinishPfx
	focusFinishPassword
	focusFinishChain
	focusFinishRoot
	focusFinishFinish
	focusFinishBack
	focusFinishTotal
)
const (
	fieldPfx      = 0
	fieldPassword = 1
)

type FinishModel struct {
	certPicker       filepicker.Model
	certPickerOpened bool
	keyPicker        filepicker.Model
	keyPickerOpened  bool
	selectedCert     string
	selectedKey      string
	keyWasGuessed    bool
	pfxInput         textinput.Model
	password         textinput.Model
	includeChain     bool
	includeRoot      bool
	focused          finishFocus
	editing          bool
	err              error
	success          bool
	statusMsg        string
}

func NewFinishModel() *FinishModel {
	certPicker := filepicker.New()
	//certPicker.AllowedTypes = []string{".crt", ".cer", ".pem"}
	//certPicker.AllowedTypes = nil

	certPicker.CurrentDirectory = "."
	certPicker.AutoHeight = false
	certPicker.SetHeight(0)
	//certPicker.SetHeight(10)

	keyPicker := filepicker.New()
	//keyPicker.AllowedTypes = []string{".key", ".pem"}
	//keyPicker.AllowedTypes = nil
	keyPicker.CurrentDirectory = "."
	keyPicker.AutoHeight = false
	keyPicker.SetHeight(0)
	//keyPicker.SetHeight(10)

	pfxInput := textinput.New()
	pfxInput.Placeholder = "output.pfx"
	pfxInput.CharLimit = 256
	pfxInput.Width = FieldWidth
	pfxInput.Prompt = ""

	password := textinput.New()
	password.Placeholder = "optional"
	password.EchoMode = textinput.EchoPassword
	password.Prompt = ""

	m := FinishModel{
		certPicker:   certPicker,
		keyPicker:    keyPicker,
		pfxInput:     pfxInput,
		password:     password,
		includeChain: false,
		includeRoot:  false,
		editing:      false,
		focused:      focusFinishCert,
	}

	//m.updateFocus()
	return &m
}

func (m *FinishModel) updateFocus() {
	m.pfxInput.Blur()
	m.password.Blur()

	if m.editing {
		switch m.focused {
		case focusFinishPfx:
			m.pfxInput.Focus()
		case focusFinishPassword:
			m.password.Focus()
		}
	}
}

func (m *FinishModel) Init() tea.Cmd {
	return nil
}

func (m *FinishModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Optional: implement window responsiveness
		m.certPicker, cmd = m.certPicker.Update(msg)
		cmds = append(cmds, cmd)
		m.keyPicker, cmd = m.keyPicker.Update(msg)
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			if !m.editing {
				// add the back handler to be processed
				cmds = append(cmds, tea.Sequence(func() tea.Msg {
					_, cmd := m.handleBack()
					return tea.Cmd(cmd)()
				}))
			}
			// handle escape when either picker are opened
			if m.certPickerOpened {
				m.certPickerOpened = false
				// handle closing cert picker, or?
			}
			if m.keyPickerOpened {
				m.keyPickerOpened = false
			}

		case "up", "k":
			if !m.certPickerOpened && !m.keyPickerOpened {
				if !m.editing {
					m.focused--
					if m.focused < 0 {
						m.focused = focusFinishTotal - 1
					}
					m.updateFocus()
				}
			}
			// don't return yet, let messages bubble down
			//return m, nil

		case "down", "j":
			if !m.certPickerOpened && !m.keyPickerOpened {
				if !m.editing {
					m.focused = (m.focused + 1) % focusFinishTotal
					m.updateFocus()
				}
			}
			// don't return yet, let messages bubble down
			//return m, nil

		case "enter":
			if m.editing {
				m.editing = false
				m.updateFocus()
				// not sure if this should be here, it wasn't before
				cmds = append(cmds, textinput.Blink)
			} else {
				// neither of the pickers are opened
				if !m.certPickerOpened && !m.keyPickerOpened {
					switch m.focused {
					case focusFinishCert:
						m.certPickerOpened = true
						m.certPicker.SetHeight(10)
						cmd := m.certPicker.Init()
						cmds = append(cmds, cmd)
						//m.updateFocus()
					case focusFinishKey:
						m.keyPickerOpened = true
						m.keyPicker.SetHeight(10)
						m.keyPicker.AutoHeight = true
						cmd := m.keyPicker.Init()
						cmds = append(cmds, cmd)
						//m.updateFocus()
					case focusFinishPfx:
						m.editing = true
						m.updateFocus()
						var cmd tea.Cmd
						m.pfxInput, cmd = m.pfxInput.Update(msg)
						cmds = append(cmds, textinput.Blink)
						cmds = append(cmds, cmd)
					case focusFinishPassword:
						m.editing = true
						m.updateFocus()
						var cmd tea.Cmd
						m.password, cmd = m.password.Update(msg)
						cmds = append(cmds, textinput.Blink)
						cmds = append(cmds, cmd)
					case focusFinishChain:
						m.includeChain = !m.includeChain
						if !m.includeChain {
							m.includeRoot = false
						}
					case focusFinishRoot:
						if m.includeChain {
							m.includeRoot = !m.includeRoot
						}
					case focusFinishFinish:
						/*
							cmds = append(cmds, tea.Sequence(func() tea.Msg {
								_, cmd := m.handleFinish()
								return tea.Cmd(cmd)()
							}))
						*/
						// For the finish option, just return the finish handler
						return m.handleFinish()

					case focusFinishBack:
						cmds = append(cmds, tea.Sequence(func() tea.Msg {
							_, cmd := m.handleBack()
							return tea.Cmd(cmd)()
						}))
					}
				} else {
					// cert picker is opened, so we need to select a file...
					if m.certPickerOpened {
						if didSelect, path := m.certPicker.DidSelectFile(msg); didSelect {
							m.selectedCert = path
							m.certPicker.SetHeight(0)
							m.certPickerOpened = false
						}
					}
					// key picker is opened, so we need to select a file...
					if m.keyPickerOpened {
						if didSelect, path := m.keyPicker.DidSelectFile(msg); didSelect {
							m.selectedKey = path
							m.keyPicker.SetHeight(0)
							m.keyPickerOpened = false
						}
					}
				}
			}
		}
	}

	m.certPicker, cmd = m.certPicker.Update(msg)
	cmds = append(cmds, cmd)

	m.keyPicker, cmd = m.keyPicker.Update(msg)
	cmds = append(cmds, cmd)

	// handle editing inputs to fields (pfx and pasword)
	if m.editing {
		switch m.focused {
		case focusFinishPfx:
			m.pfxInput, cmd = m.pfxInput.Update(msg)
		case focusFinishPassword:
			m.password, cmd = m.password.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	certPicked, certFile := m.certPicker.DidSelectFile(msg)
	if certPicked && m.selectedCert == "" {
		m.selectedCert = certFile

		guessKey := strings.TrimSuffix(certFile, filepath.Ext(certFile)) + ".key"
		if _, err := os.Stat(guessKey); err == nil {
			m.selectedKey = guessKey
			m.keyWasGuessed = true
		}
	}

	keyPicked, keyFile := m.keyPicker.DidSelectFile(msg)
	if keyPicked && m.selectedKey == "" {
		m.selectedKey = keyFile
		m.keyWasGuessed = false
	}

	return m, tea.Batch(cmds...)
}

func (m *FinishModel) renderField(label string, inp textinput.Model, idx finishFocus) string {
	focused := m.focused == idx
	style := BlurredStyle
	if focused {
		style = FocusedStyle
	}

	// if editing, show the input.View()
	if m.editing && focused {
		return fmt.Sprintf("%s %s\n", LabelStyle.Render(label+":"),
			inp.View())
	}

	// else show label and value with cursor indicator if focused
	cursor := " "
	if focused {
		cursor = ">"
	}
	val := inp.Value()
	if val == "" {
		val = "<none>"
	}
	return fmt.Sprintf("%s %s %s\n", CursorStyle.Render(cursor), style.Render(label+":"), style.Render(val))
}

// TODO - Need to fix the unstyled text labels for the file pickers (they are showing WHITE)
func (m *FinishModel) View() string {
	var b strings.Builder

	b.WriteString("Use ↑/↓ (k/j) to navigate, Enter to edit/select\n\n")
	cpMsg := "Choose cert file"
	cpCursor := " "
	if m.certPickerOpened {
		cpMsg = "Cert Picker Open"
	}
	cStyle := BlurredStyle //LabelStyle
	if m.focused == focusFinishCert {
		cStyle = FocusedStyle
		cpCursor = ">"
	}
	if m.selectedCert == "" {
		b.WriteString(fmt.Sprintf("%s %s\n", CursorStyle.Render(cpCursor), cStyle.Render(cpMsg)))
	} else {
		b.WriteString(fmt.Sprintf("%s %s\n", CursorStyle.Render(cpCursor), cStyle.Render("["+m.selectedCert+"]")))
	}
	// maybe only show this if certPickerOpened
	if m.certPickerOpened {
		b.WriteString(m.certPicker.View() + "\n")
	}

	kpMsg := "Choose private key file"
	kpCursor := " "
	if m.keyPickerOpened {
		kpMsg = "Key Picker Open"
	}
	kStyle := BlurredStyle //LabelStyle
	if m.focused == focusFinishKey {
		kStyle = FocusedStyle
		kpCursor = ">"
	}
	if m.selectedKey == "" {
		b.WriteString(fmt.Sprintf("%s %s\n", CursorStyle.Render(kpCursor), kStyle.Render(kpMsg)))
	} else {
		b.WriteString(fmt.Sprintf("%s %s\n", CursorStyle.Render(kpCursor), kStyle.Render("["+m.selectedKey+"]")))
	}
	// maybe only show this if keyPickerOpened
	if m.keyPickerOpened {
		b.WriteString(m.keyPicker.View() + "\n")
	}

	// **************** PFX field *********************
	b.WriteString(m.renderField("Pfx File", m.pfxInput, focusFinishPfx))

	// **************** Password field *********************
	b.WriteString(m.renderField("Password", m.password, focusFinishPassword))

	chainLabel := "[ ] Include certificate chain"
	if m.includeChain {
		chainLabel = "[✓] Include certificate chain"
	}
	if m.focused == focusFinishChain {
		b.WriteString(FocusedStyle.Render(chainLabel) + "\n")
	} else {
		b.WriteString(BlurredStyle.Render(chainLabel) + "\n")
	}

	rootLabel := "[ ] Include root certificate"
	if m.includeRoot {
		rootLabel = "[✓] Include root certificate"
	}
	if !m.includeChain {
		rootLabel += " (requires chain)"
	}
	if m.focused == focusFinishRoot {
		b.WriteString(FocusedStyle.Render(rootLabel) + "\n")
	} else {
		b.WriteString(BlurredStyle.Render(rootLabel) + "\n")
	}

	if m.focused == focusFinishFinish {
		b.WriteString(ButtonFocusedStyle.Render("Finish") + "\n")
	} else {
		b.WriteString(ButtonBlurredStyle.Render("Finish") + "\n")
	}

	if m.focused == focusFinishBack {
		b.WriteString(ButtonFocusedStyle.Render("Back to Menu") + "\n")
	} else {
		b.WriteString(ButtonBlurredStyle.Render("Back to Menu") + "\n")
	}
	if m.statusMsg != "" {
		if m.success {
			b.WriteString(MsgSuccessStyle.Render("\n" + m.statusMsg + "\n"))
		} else {
			b.WriteString(MsgErrorStyle.Render("\n" + m.statusMsg + "\n"))
		}
	}

	return b.String()
}

// handleBack sends a Quit or a custom Msg to parent model to return to main menu
func (m *FinishModel) handleBack() (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		return BackToMenuMsg{}
	}
}

func wrapText(s string, limit int) string {
	if len(s) <= limit {
		return s
	}

	var b strings.Builder
	words := strings.Fields(s)
	lineLen := 0

	for _, w := range words {
		if lineLen+len(w)+1 > limit {
			b.WriteString("\n")
			lineLen = 0
		} else if lineLen > 0 {
			b.WriteString(" ")
			lineLen++
		}
		b.WriteString(w)
		lineLen += len(w)
	}

	return b.String()
}

func (m *FinishModel) handleFinish() (tea.Model, tea.Cmd) {
	opts := options.FinishOptions{
		Certificate: m.selectedCert,
		Key:         m.selectedKey,
		PfxFile:     m.pfxInput.Value(),
		Password:    m.password.Value(),
		Chain:       m.includeChain,
		IncludeRoot: m.includeRoot,
		Verbose:     false,
	}
	if err := opts.Validate(); err != nil {
		m.err = err
		m.statusMsg = fmt.Sprintf("Error validating options: %s", err)
		return m, func() tea.Msg {
			return ResultErrorMsg{err:err}
		}
	}

	svc := certsvc.New()
	result, err := svc.FinishCSR(opts)
	if err != nil {
		m.err = err
		return m, func() tea.Msg {
			return ResultErrorMsg{err:err}
		}
		
	}
	if opts.PfxFile != "" {
		// TODO - we should use the full path of the cert's directory instead of just name
		result.FileName = opts.PfxFile
	}
	outputLines, err := svc.SavePFXdto(result)
	if err != nil {
		m.err = err
		return m, func() tea.Msg {
			return ResultErrorMsg{err: err}
		}
	}
	m.success = true
	/* this debug code will add arguments to output
	wrappedLines := wrapText( fmt.Sprintf("%+v",opts),100)
	outputLines = append(outputLines, "Arguments given: ")
	outputLines = append(outputLines, wrappedLines)
	*/
	return m, func() tea.Msg {
		return SuccessMsg{logs: outputLines}
	}
}
