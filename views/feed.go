package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/euklides/cyberspace-cli/api"
	"github.com/euklides/cyberspace-cli/models"
	"github.com/euklides/cyberspace-cli/styles"
)

// PostsLoadedMsg is sent when posts are fetched
type PostsLoadedMsg struct {
	Posts []models.Post
}

// PostsErrorMsg is sent when fetching posts fails
type PostsErrorMsg struct {
	Err error
}

// OpenPostMsg is sent when user wants to view a post
type OpenPostMsg struct {
	Post models.Post
}

// FeedModel is the post feed screen
type FeedModel struct {
	posts   []models.Post
	cursor  int
	offset  int
	loading bool
	spinner spinner.Model
	err     error
	client  *api.FirestoreClient
	width   int
	height  int
}

// NewFeedModel creates a new feed screen
func NewFeedModel(projectID, idToken string) FeedModel {
	return FeedModel{
		client:  api.NewFirestoreClient(projectID, idToken),
		spinner: NewSpinner(),
		loading: true,
	}
}

func (m FeedModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchPosts())
}

func (m FeedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.posts)-1 {
				m.cursor++
				m.adjustScroll()
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.adjustScroll()
			}
		case "g":
			m.cursor = 0
			m.offset = 0
		case "G":
			m.cursor = len(m.posts) - 1
			m.adjustScroll()
		case "r":
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.fetchPosts())
		case "enter":
			if len(m.posts) > 0 && m.cursor < len(m.posts) {
				return m, func() tea.Msg {
					return OpenPostMsg{Post: m.posts[m.cursor]}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.adjustScroll()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case PostsLoadedMsg:
		m.loading = false
		m.posts = msg.Posts
		m.cursor = 0
		m.offset = 0

	case PostsErrorMsg:
		m.loading = false
		m.err = msg.Err
	}

	return m, nil
}

func (m FeedModel) View() string {
	if m.loading {
		return RenderLoading(m.spinner, "Loading posts...", m.width, m.height)
	}

	if m.err != nil {
		return RenderError(m.err, "Press 'r' to retry, 'q' to quit", m.width, m.height)
	}

	if len(m.posts) == 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			"No posts found. Press 'r' to refresh.")
	}

	var b strings.Builder

	// Header
	b.WriteString(RenderHeader("CYBERSPACE FEED", "j/k:nav  enter:open  r:refresh  q:quit", m.width))

	// Calculate visible posts
	visiblePosts := m.visiblePostCount()

	// Render posts
	for i := m.offset; i < len(m.posts) && i < m.offset+visiblePosts; i++ {
		b.WriteString(m.renderPost(m.posts[i], i == m.cursor))
		b.WriteString("\n")
	}

	// Footer
	b.WriteString(styles.Footer.Render(fmt.Sprintf("Post %d of %d", m.cursor+1, len(m.posts))))

	return b.String()
}

func (m FeedModel) visiblePostCount() int {
	visibleHeight := m.height - 4
	postHeight := 5
	count := visibleHeight / postHeight
	if count < 1 {
		return 1
	}
	return count
}

func (m *FeedModel) adjustScroll() {
	visiblePosts := m.visiblePostCount()

	if m.cursor >= m.offset+visiblePosts {
		m.offset = m.cursor - visiblePosts + 1
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
}

func (m FeedModel) renderPost(p models.Post, selected bool) string {
	width := Min(m.width-2, 78)
	if width < 1 {
		width = 78
	}

	baseStyle := lipgloss.NewStyle().Padding(0, 1).Width(width)

	if selected {
		baseStyle = baseStyle.
			Background(styles.ColorBgSelect).
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(styles.ColorPrimary)
	}

	header := FormatAuthor(p.AuthorUsername, p.CreatedAt)
	content := styles.Content.Render(Truncate(StripMarkdown(p.Content), 140))
	stats := FormatStats(p.RepliesCount, p.BookmarksCount)

	if len(p.Topics) > 0 {
		stats = stats + "  " + FormatTopics(p.Topics)
	}

	return baseStyle.Render(header + "\n" + content + "\n" + stats)
}

func (m FeedModel) fetchPosts() tea.Cmd {
	return func() tea.Msg {
		posts, err := m.client.FetchPosts(20)
		if err != nil {
			return PostsErrorMsg{Err: err}
		}
		return PostsLoadedMsg{Posts: posts}
	}
}
