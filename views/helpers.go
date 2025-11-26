package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"

	"github.com/euklides/cyberspace-cli/styles"
)

// TimeAgo formats a time as a relative string (e.g., "5m", "2h", "3d")
func TimeAgo(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		m := int(d.Minutes())
		return fmt.Sprintf("%dm", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf("%dh", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%dd", days)
	default:
		return t.Format("Jan 2")
	}
}

// Truncate shortens a string to max length with ellipsis
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// StripMarkdown removes basic markdown formatting for plain text display
func StripMarkdown(s string) string {
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "__", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "#", "")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

// Min returns the smaller of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the larger of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Clamp restricts a value to a range
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// SafeWidth returns a width that's safe for rendering (minimum 1)
func SafeWidth(width, defaultWidth int) int {
	if width < 1 {
		return defaultWidth
	}
	return width
}

// NewSpinner creates a consistently styled spinner
func NewSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Spinner
	return s
}

// RenderHeader renders a consistent header with title and help text
func RenderHeader(title, help string, width int) string {
	var b strings.Builder

	b.WriteString(styles.Header.Render(title))
	b.WriteString("  ")
	b.WriteString(styles.Help.Render(help))
	b.WriteString("\n")

	dividerWidth := Min(width, 80)
	if dividerWidth < 1 {
		dividerWidth = 80
	}
	b.WriteString(styles.Divider(dividerWidth))
	b.WriteString("\n")

	return b.String()
}

// RenderError renders an error message centered on screen
func RenderError(err error, hint string, width, height int) string {
	msg := styles.Error.Render(fmt.Sprintf("Error: %s", err.Error()))
	if hint != "" {
		msg += "\n\n" + hint
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, msg)
}

// RenderLoading renders a loading message centered on screen
func RenderLoading(s spinner.Model, message string, width, height int) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
		s.View()+" "+message)
}

// FormatTopics formats a topics array as a hashtag string
func FormatTopics(topics []string) string {
	if len(topics) == 0 {
		return ""
	}
	return styles.Topic.Render("#" + strings.Join(topics, " #"))
}

// FormatStats formats reply and bookmark counts
func FormatStats(replies, bookmarks int) string {
	return styles.Stats.Render(fmt.Sprintf("↩ %d  ★ %d", replies, bookmarks))
}

// FormatAuthor formats username and timestamp
func FormatAuthor(username string, createdAt time.Time) string {
	return fmt.Sprintf("%s %s",
		styles.Username.Render("@"+username),
		styles.Timestamp.Render("· "+TimeAgo(createdAt)))
}
