package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/euklides/cyberspace-cli/api"
	"github.com/euklides/cyberspace-cli/styles"
)

// LoginSuccessMsg is sent when login succeeds
type LoginSuccessMsg struct {
	IDToken      string
	RefreshToken string
	UserID       string
	Email        string
}

// LoginErrorMsg is sent when login fails
type LoginErrorMsg struct {
	Err error
}

// LoginModel is the login screen
type LoginModel struct {
	emailInput    textinput.Model
	passwordInput textinput.Model
	focusIndex    int
	loading       bool
	err           error
	apiKey        string
	width         int
	height        int
}

// NewLoginModel creates a new login screen
func NewLoginModel(apiKey string) LoginModel {
	ei := textinput.New()
	ei.Placeholder = "email@example.com"
	ei.Focus()
	ei.CharLimit = 64
	ei.Width = 30

	pi := textinput.New()
	pi.Placeholder = "password"
	pi.EchoMode = textinput.EchoPassword
	pi.EchoCharacter = '•'
	pi.CharLimit = 64
	pi.Width = 30

	return LoginModel{
		emailInput:    ei,
		passwordInput: pi,
		focusIndex:    0,
		apiKey:        apiKey,
	}
}

func (m LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % 2
			return m, m.updateFocus()
		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + 2) % 2
			return m, m.updateFocus()
		case "enter":
			if m.focusIndex == 1 || (m.emailInput.Value() != "" && m.passwordInput.Value() != "") {
				m.loading = true
				m.err = nil
				return m, m.attemptLogin()
			}
			if m.focusIndex == 0 {
				m.focusIndex = 1
				return m, m.updateFocus()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case LoginSuccessMsg:
		m.loading = false
		return m, nil

	case LoginErrorMsg:
		m.loading = false
		m.err = msg.Err
		return m, nil
	}

	// Update the focused input
	var cmd tea.Cmd
	if m.focusIndex == 0 {
		m.emailInput, cmd = m.emailInput.Update(msg)
	} else {
		m.passwordInput, cmd = m.passwordInput.Update(msg)
	}

	return m, cmd
}

func (m LoginModel) View() string {
	var form strings.Builder

	// Title
	titleStyle := styles.Title.Copy().MarginBottom(1)
	form.WriteString(titleStyle.Render("CYBERSPACE"))
	form.WriteString("\n\n")

	// Email field
	form.WriteString(styles.Label.Render("Email"))
	form.WriteString("\n")
	form.WriteString(m.emailInput.View())
	form.WriteString("\n\n")

	// Password field
	form.WriteString(styles.Label.Render("Password"))
	form.WriteString("\n")
	form.WriteString(m.passwordInput.View())
	form.WriteString("\n\n")

	// Status
	if m.loading {
		form.WriteString("Signing in...")
	} else if m.err != nil {
		form.WriteString(styles.Error.Render(fmt.Sprintf("Error: %s", m.err.Error())))
	} else {
		form.WriteString(styles.Help.Render("Press Enter to sign in"))
	}

	// Wrap in box
	box := styles.Box.Render(form.String())

	// Center on screen
	if m.width > 0 && m.height > 0 {
		box = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}

	return box
}

func (m *LoginModel) updateFocus() tea.Cmd {
	if m.focusIndex == 0 {
		m.passwordInput.Blur()
		return m.emailInput.Focus()
	}
	m.emailInput.Blur()
	return m.passwordInput.Focus()
}

func (m LoginModel) attemptLogin() tea.Cmd {
	return func() tea.Msg {
		resp, err := api.SignIn(m.emailInput.Value(), m.passwordInput.Value(), m.apiKey)
		if err != nil {
			return LoginErrorMsg{Err: err}
		}

		return LoginSuccessMsg{
			IDToken:      resp.IDToken,
			RefreshToken: resp.RefreshToken,
			UserID:       resp.LocalID,
			Email:        resp.Email,
		}
	}
}
