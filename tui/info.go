// tui/info.go
package tui

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"ssl-tools/internal/certformat"
	"ssl-tools/internal/certsvc"
	"ssl-tools/internal/options"
	"strings"
)

type InfoModel struct {
	filePicker       filepicker.Model
	filePickerOpened bool
	selectedFile     string
	hostInput        textinput.Model
	addrInput        textinput.Model
	summary          bool
	passwordInput    textinput.Model
	focused          infoFocus
	editing          bool
	err              error
	success          bool
	statusMsg        string
	viewport         viewport.Model
	viewReady        bool
	viewWidth        int
	viewHeight       int
	scrolling        bool
}

// focus handlers
type infoFocus int

const (
	focusInfoPicker infoFocus = iota
	focusInfoHost
	focusInfoAddr
	focusInfoSummary
	focusInfoPassword
	focusInfoAction
	focusInfoBack
	focusInfoViewport
	focusInfoTotal
)

func NewInfoModel() *InfoModel {
	filePicker := filepicker.New()
	filePicker.CurrentDirectory = "."
	filePicker.AutoHeight = false
	filePicker.SetHeight(0)

	hostInput := textinput.New()
	hostInput.Placeholder = "hostname"
	hostInput.Width = FieldWidth
	hostInput.CharLimit = 256
	hostInput.Prompt = ""

	addrInput := textinput.New()
	addrInput.Placeholder = "address"
	addrInput.Width = FieldWidth
	addrInput.CharLimit = 256
	addrInput.Prompt = ""

	password := textinput.New()
	password.Placeholder = "optional"
	password.EchoMode = textinput.EchoPassword
	password.Width = FieldWidth
	password.CharLimit = 256
	password.Prompt = ""
	m := InfoModel{
		filePicker:    filePicker,
		hostInput:     hostInput,
		addrInput:     addrInput,
		passwordInput: password,
		editing:       false,
		focused:       focusInfoPicker,
	}
	return &m
}

func (m *InfoModel) Init() tea.Cmd {
	return nil
}

// handle text input focus
func (m *InfoModel) updateFocus() {
	m.hostInput.Blur()
	m.addrInput.Blur()

	if m.editing {
		switch m.focused {
		case focusInfoHost:
			m.hostInput.Focus()
		case focusInfoAddr:
			m.addrInput.Focus()
		}
	}
}

func (m *InfoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.viewReady {
			m.viewport = viewport.New(msg.Width, msg.Height-20)
			m.viewWidth = msg.Width
			m.viewHeight = msg.Height
			//m.viewport.SetContent("sample viewport")
			//m.statusMsg = "sample viewport"
			m.viewReady = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 20
		}
		// not sure if this is actually needed but will keep for now
		//m.filePicker, cmd = m.filePicker.Update(msg)
		//cmds = append(cmds, cmd)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if !m.scrolling {
				// handle escape/back when picker is opened
				if m.filePickerOpened {
					m.filePickerOpened = false
					m.selectedFile = ""
				} else {
					cmds = append(cmds, tea.Sequence(func() tea.Msg {
						_, cmd := m.handleBack()
						return tea.Cmd(cmd)()
					}))
				}
			} else {
				// set scrolling false and reset focus
				m.scrolling = false
				m.focused = focusInfoTotal - 1
			}
		case "q":
			if !m.scrolling {
				if !m.editing {
					if m.filePickerOpened {
						m.filePickerOpened = false
						m.selectedFile = ""
					} else {
						cmds = append(cmds, tea.Sequence(func() tea.Msg {
							_, cmd := m.handleBack()
							return tea.Cmd(cmd)()
						}))
					}
				}
			}
		case "up", "k":
			if !m.scrolling {
				if !m.filePickerOpened {
					if !m.editing {
						m.focused--
						if m.focused < 0 {
							m.focused = focusInfoTotal - 1
						}
						if m.focused == focusInfoViewport {
							if strings.TrimSpace(m.viewport.View()) != "" {
								m.scrolling = true
							} else {
								m.focused = focusInfoTotal - 2
							}
						}
						m.updateFocus()
					}
				}
			}
		case "down", "j":
			if !m.scrolling {
				if !m.filePickerOpened {
					if !m.editing {
						m.focused = (m.focused + 1) % focusInfoTotal
						if m.focused == focusInfoViewport {
							if strings.TrimSpace(m.viewport.View()) != "" {
								m.scrolling = true
							} else {
								m.focused = (m.focused + 1) % focusInfoTotal
							}
						}
						m.updateFocus()
					}
				}
			}
		case "enter":
			if !m.scrolling {
				if m.editing {
					m.editing = false
					m.updateFocus()
					// lets blink after focus?
					cmds = append(cmds, textinput.Blink)
				} else {
					// if the picker is not open
					if !m.filePickerOpened {
						switch m.focused {
						case focusInfoPicker:
							m.filePickerOpened = true
							m.filePicker.SetHeight(12)
							cmd := m.filePicker.Init()
							cmds = append(cmds, cmd)
						case focusInfoHost:
							m.editing = true
							m.updateFocus()
							var cmd tea.Cmd
							m.hostInput, cmd = m.hostInput.Update(msg)
							cmds = append(cmds, textinput.Blink)
							cmds = append(cmds, cmd)
						case focusInfoAddr:
							m.editing = true
							m.updateFocus()
							var cmd tea.Cmd
							m.addrInput, cmd = m.addrInput.Update(msg)
							cmds = append(cmds, textinput.Blink)
							cmds = append(cmds, cmd)
						case focusInfoSummary:
							// toggle summary option
							m.summary = !m.summary
						case focusInfoPassword:
							m.editing = true
							m.updateFocus()
							var cmd tea.Cmd
							m.passwordInput, cmd = m.passwordInput.Update(msg)
							cmds = append(cmds, textinput.Blink)
							cmds = append(cmds, cmd)
						case focusInfoAction:
							// here we will do the main action
							// will likely be return m.handleAction()
							//return m, nil
							return m.handleAction()
						case focusInfoBack:
							cmds = append(cmds, tea.Sequence(func() tea.Msg {
								_, cmd := m.handleBack()
								return tea.Cmd(cmd)()
							}))
						}
					} else {
						// file picker is opened so we need to select a file...
						if m.filePickerOpened {
							if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
								m.selectedFile = path
								m.filePicker.SetHeight(0)
								m.filePickerOpened = false
							}
						}
					}
				}
			}
		}
	}

	if !m.scrolling {
		// outside of main (msg) switch block
		m.filePicker, cmd = m.filePicker.Update(msg)
		cmds = append(cmds, cmd)
	}

	// handle editing inputs to fields (host, addr, and password)
	if !m.scrolling && m.editing {
		switch m.focused {
		case focusInfoHost:
			m.hostInput, cmd = m.hostInput.Update(msg)
		case focusInfoAddr:
			m.addrInput, cmd = m.addrInput.Update(msg)
		case focusInfoPassword:
			m.passwordInput, cmd = m.passwordInput.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	filePicked, file := m.filePicker.DidSelectFile(msg)
	if filePicked && m.selectedFile == "" {
		m.selectedFile = file
	}

	if m.scrolling {
		// give control to viewport for scrolling
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *InfoModel) handleBack() (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		return BackToMenuMsg{}
	}
}

func (m *InfoModel) handleAction() (tea.Model, tea.Cmd) {
	opts := options.InfoOptions{
		Password:     m.passwordInput.Value(),
		ShortSummary: m.summary,
		//Certificates []string
		//URLs []string
		//Hosts map[string]string
		//CSR string
	}
	m.success = false
	// need to check host/addr and set the appropriate option
	if m.hostInput.Value() != "" && m.addrInput.Value() == "" {
		// if we have a host but no addr we have URL
		opts.URLs = []string{m.hostInput.Value()}
	}
	if m.hostInput.Value() != "" && m.addrInput.Value() != "" {
		// we have a host and addr, so add a host map
		opts.Hosts = make(map[string]string)
		opts.Hosts[m.hostInput.Value()] = m.addrInput.Value()
	}
	// also need to check file and check if it is a cert or csr
	if m.selectedFile != "" {
		format := certformat.CertificateFormat.Detect(m.selectedFile)
		switch format {
		case certformat.DER:
			opts.Certificates = []string{m.selectedFile}
		case certformat.PEM:
			opts.Certificates = []string{m.selectedFile}
		default:
			opts.CSR = m.selectedFile
		}
	}
	if err := opts.Validate(); err != nil {
		m.err = err
		m.statusMsg = fmt.Sprintf("Error validating options: %v", errors.Unwrap(err))
		return m, func() tea.Msg {
			return ResultErrorMsg{err: err}
		}
	}
	svc := certsvc.New()
	//outputLines := []string{"Test info result output"}
	outputLines, err := svc.GetInfo(opts)
	if err != nil {
		m.err = err
		return m, func() tea.Msg {
			return ResultErrorMsg{err: err}
		}
	}
	m.success = true
	m.scrolling = true
	return m, func() tea.Msg {
		//m.statusMsg = "content delivered"
		return SuccessMsg{logs: outputLines}
	}
}

func (m *InfoModel) renderField(label string, inp textinput.Model, idx infoFocus) string {
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

	// else show label and value with curstor indicator if focused
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

func (m *InfoModel) View() string {
	var b strings.Builder

	b.WriteString("Use ↑/↓ (k/j) to navigate, Enter to edit/select\n\n")
	fpMsg := "Choose file"
	fpCursor := " "
	if m.filePickerOpened {
		fpMsg = "File Picker Opened"
	}
	cStyle := BlurredStyle
	if m.focused == focusInfoPicker {
		cStyle = FocusedStyle
		fpCursor = ">"
	}
	if m.selectedFile == "" {
		b.WriteString(fmt.Sprintf("%s %s\n", CursorStyle.Render(fpCursor), cStyle.Render(fpMsg)))
	} else {
		b.WriteString(fmt.Sprintf("%s %s\n", CursorStyle.Render(fpCursor), cStyle.Render("["+m.selectedFile+"]")))
	}
	if m.filePickerOpened {
		b.WriteString(m.filePicker.View() + "\n")
	}

	// host input field
	b.WriteString(m.renderField("Host", m.hostInput, focusInfoHost))

	// addr input field
	b.WriteString(m.renderField("Address", m.addrInput, focusInfoAddr))

	// summary checkbox
	summaryLabel := "[ ] Summary view"
	if m.summary {
		summaryLabel = "[✓] Summary view"
	}
	if m.focused == focusInfoSummary {
		b.WriteString(FocusedStyle.Render(summaryLabel) + "\n")
	} else {
		b.WriteString(BlurredStyle.Render(summaryLabel) + "\n")
	}

	// password input field
	b.WriteString(m.renderField("Password", m.passwordInput, focusInfoPassword))

	// buttons
	if m.focused == focusInfoAction {
		b.WriteString(ButtonFocusedStyle.Render("Get Info") + "\n")
	} else {
		b.WriteString(ButtonBlurredStyle.Render("Get Info") + "\n")
	}

	if m.focused == focusInfoBack {
		b.WriteString(ButtonFocusedStyle.Render("Back to Menu") + "\n")
	} else {
		b.WriteString(ButtonBlurredStyle.Render("Back to Menu") + "\n")
	}

	vp := m.viewport.View()
	if vp != "" {
		if m.focused == focusInfoViewport || m.scrolling {
			//b.WriteString(MsgSuccessStyle.Render("\n" + m.statusMsg + "\n"))
			b.WriteString(LabelStyle.Render(m.viewport.View()))
		} else {
			//b.WriteString(MsgErrorStyle.Render("\n" + m.statusMsg + "\n"))
			b.WriteString(BlurredStyle.Render(m.viewport.View()))
		}
	}
	return b.String()
}
