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

// PostDetailLoadedMsg is sent when post and replies are loaded
type PostDetailLoadedMsg struct {
	Post    models.Post
	Replies []models.Reply
}

// PostDetailErrorMsg is sent when loading fails
type PostDetailErrorMsg struct {
	Err error
}

// BackToFeedMsg is sent when user wants to go back
type BackToFeedMsg struct{}

// PostDetailModel is the post detail screen
type PostDetailModel struct {
	post    models.Post
	replies []models.Reply
	scroll  int
	loading bool
	spinner spinner.Model
	err     error
	client  *api.FirestoreClient
	postID  string
	width   int
	height  int
}

// NewPostDetailModel creates a new post detail screen
func NewPostDetailModel(projectID, idToken, postID string) PostDetailModel {
	return PostDetailModel{
		client:  api.NewFirestoreClient(projectID, idToken),
		postID:  postID,
		spinner: NewSpinner(),
		loading: true,
	}
}

// NewPostDetailModelWithPost creates a detail screen with post already loaded
func NewPostDetailModelWithPost(projectID, idToken string, post models.Post) PostDetailModel {
	return PostDetailModel{
		client:  api.NewFirestoreClient(projectID, idToken),
		postID:  post.ID,
		post:    post,
		spinner: NewSpinner(),
		loading: true,
	}
}

func (m PostDetailModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetchPostAndReplies())
}

func (m PostDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc", "b", "backspace":
			return m, func() tea.Msg { return BackToFeedMsg{} }
		case "j", "down":
			m.scroll++
			m.clampScroll()
		case "k", "up":
			m.scroll--
			m.clampScroll()
		case "g":
			m.scroll = 0
		case "G":
			m.scroll = m.maxScroll()
		case "r":
			m.loading = true
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, m.fetchPostAndReplies())
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.clampScroll()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case PostDetailLoadedMsg:
		m.loading = false
		m.post = msg.Post
		m.replies = msg.Replies

	case PostDetailErrorMsg:
		m.loading = false
		m.err = msg.Err
	}

	return m, nil
}

func (m PostDetailModel) View() string {
	if m.loading && m.post.ID == "" {
		return RenderLoading(m.spinner, "Loading post...", m.width, m.height)
	}

	if m.err != nil {
		return RenderError(m.err, "Press 'b' to go back", m.width, m.height)
	}

	var b strings.Builder

	// Header
	b.WriteString(RenderHeader("POST", "b:back  j/k:scroll  r:refresh  q:quit", m.width))
	b.WriteString("\n")

	// Build content
	content := m.buildContent()
	lines := strings.Split(content, "\n")

	// Apply scroll
	visibleLines := m.height - 4
	start := m.scroll
	end := Min(start+visibleLines, len(lines))

	if start < len(lines) {
		for i := start; i < end; i++ {
			b.WriteString(lines[i])
			b.WriteString("\n")
		}
	}

	// Scroll indicator
	if len(lines) > visibleLines {
		b.WriteString("\n")
		b.WriteString(styles.Help.Render(fmt.Sprintf("Line %d-%d of %d", start+1, end, len(lines))))
	}

	return b.String()
}

func (m PostDetailModel) buildContent() string {
	var b strings.Builder

	contentWidth := Min(m.width-4, 76)
	if contentWidth < 1 {
		contentWidth = 76
	}

	// Author and time
	b.WriteString(FormatAuthor(m.post.AuthorUsername, m.post.CreatedAt))
	b.WriteString("\n\n")

	// Full content
	contentStyle := styles.Content.Copy().Width(contentWidth)
	b.WriteString(contentStyle.Render(m.post.Content))
	b.WriteString("\n\n")

	// Stats
	stats := fmt.Sprintf("↩ %d replies  ★ %d bookmarks", m.post.RepliesCount, m.post.BookmarksCount)
	b.WriteString(styles.Stats.Render(stats))

	// Topics
	if len(m.post.Topics) > 0 {
		b.WriteString("  ")
		b.WriteString(FormatTopics(m.post.Topics))
	}
	b.WriteString("\n")

	// Divider
	b.WriteString("\n")
	b.WriteString(styles.Divider(contentWidth))
	b.WriteString("\n")

	// Replies
	if m.loading {
		b.WriteString("\n")
		b.WriteString(m.spinner.View() + " Loading replies...")
	} else if len(m.replies) == 0 {
		b.WriteString("\n")
		b.WriteString(styles.Timestamp.Render("No replies yet"))
	} else {
		b.WriteString("\n")
		b.WriteString(styles.Stats.Render(fmt.Sprintf("REPLIES (%d)", len(m.replies))))
		b.WriteString("\n\n")

		for i, reply := range m.replies {
			b.WriteString(m.renderReply(reply))
			if i < len(m.replies)-1 {
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

func (m PostDetailModel) renderReply(r models.Reply) string {
	var b strings.Builder

	contentWidth := Min(m.width-6, 74)
	if contentWidth < 1 {
		contentWidth = 74
	}

	// Header
	b.WriteString(FormatAuthor(r.AuthorUsername, r.CreatedAt))
	b.WriteString("\n")

	// Content
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Width(contentWidth).
		PaddingLeft(2)

	b.WriteString(contentStyle.Render(r.Content))
	b.WriteString("\n")

	return b.String()
}

func (m PostDetailModel) fetchPostAndReplies() tea.Cmd {
	return func() tea.Msg {
		post := m.post
		if post.ID == "" {
			p, err := m.client.FetchPost(m.postID)
			if err != nil {
				return PostDetailErrorMsg{Err: err}
			}
			post = *p
		}

		replies, err := m.client.FetchReplies(m.postID)
		if err != nil {
			return PostDetailErrorMsg{Err: err}
		}

		return PostDetailLoadedMsg{Post: post, Replies: replies}
	}
}

func (m *PostDetailModel) clampScroll() {
	m.scroll = Clamp(m.scroll, 0, m.maxScroll())
}

func (m PostDetailModel) maxScroll() int {
	content := m.buildContent()
	lines := strings.Split(content, "\n")
	visibleLines := m.height - 4
	maxScroll := len(lines) - visibleLines
	if maxScroll < 0 {
		return 0
	}
	return maxScroll
}
