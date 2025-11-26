package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"github.com/euklides/cyberspace-cli/views"
)

// AppState represents the current screen
type AppState int

const (
	StateLogin AppState = iota
	StateFeed
	StatePostDetail
)

// Model is the main application model
type Model struct {
	state           AppState
	loginModel      views.LoginModel
	feedModel       views.FeedModel
	postDetailModel views.PostDetailModel
	config          *Config
	apiKey          string
	projectID       string
	width           int
	height          int
}

func initialModel(apiKey, projectID string, config *Config) Model {
	m := Model{
		apiKey:    apiKey,
		projectID: projectID,
		config:    config,
	}

	// If we have a saved token, go to feed
	if config != nil && config.IDToken != "" {
		m.state = StateFeed
		m.feedModel = views.NewFeedModel(projectID, config.IDToken)
	} else {
		m.state = StateLogin
		m.loginModel = views.NewLoginModel(apiKey)
	}

	return m
}

func (m Model) Init() tea.Cmd {
	switch m.state {
	case StateLogin:
		return m.loginModel.Init()
	case StateFeed:
		return m.feedModel.Init()
	case StatePostDetail:
		return m.postDetailModel.Init()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	switch m.state {
	case StateLogin:
		newLogin, cmd := m.loginModel.Update(msg)
		m.loginModel = newLogin.(views.LoginModel)

		// Check if login succeeded
		if loginMsg, ok := msg.(views.LoginSuccessMsg); ok {
			// Save config
			m.config = &Config{
				IDToken:      loginMsg.IDToken,
				RefreshToken: loginMsg.RefreshToken,
				UserID:       loginMsg.UserID,
				Email:        loginMsg.Email,
			}
			if err := SaveConfig(m.config); err != nil {
				log.Printf("Failed to save config: %v", err)
			}

			// Transition to feed view
			m.state = StateFeed
			m.feedModel = views.NewFeedModel(m.projectID, m.config.IDToken)
			return m, m.feedModel.Init()
		}

		return m, cmd

	case StateFeed:
		newFeed, cmd := m.feedModel.Update(msg)
		m.feedModel = newFeed.(views.FeedModel)

		// Check if user wants to open a post
		if openMsg, ok := msg.(views.OpenPostMsg); ok {
			m.state = StatePostDetail
			m.postDetailModel = views.NewPostDetailModelWithPost(
				m.projectID,
				m.config.IDToken,
				openMsg.Post,
			)
			return m, m.postDetailModel.Init()
		}

		return m, cmd

	case StatePostDetail:
		newDetail, cmd := m.postDetailModel.Update(msg)
		m.postDetailModel = newDetail.(views.PostDetailModel)

		// Check if user wants to go back
		if _, ok := msg.(views.BackToFeedMsg); ok {
			m.state = StateFeed
			return m, nil
		}

		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case StateLogin:
		return m.loginModel.View()
	case StateFeed:
		return m.feedModel.View()
	case StatePostDetail:
		return m.postDetailModel.View()
	default:
		return ""
	}
}

func main() {
	// Load .env file (optional - won't fail if missing)
	godotenv.Load()

	apiKey := os.Getenv("FIREBASE_API_KEY")
	projectID := os.Getenv("FIREBASE_PROJECT_ID")

	if apiKey == "" || projectID == "" {
		fmt.Println("Error: FIREBASE_API_KEY and FIREBASE_PROJECT_ID must be set")
		fmt.Println("Set them in .env file or as environment variables")
		os.Exit(1)
	}

	// Load existing config
	config, err := LoadConfig()
	if err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
	}

	// Create and run the app
	p := tea.NewProgram(
		initialModel(apiKey, projectID, config),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running app: %v\n", err)
		os.Exit(1)
	}
}
