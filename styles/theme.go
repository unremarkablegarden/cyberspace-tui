package styles

import "github.com/charmbracelet/lipgloss"

// Colors
var (
	ColorPrimary   = lipgloss.Color("214") // Pink/magenta
	ColorSecondary = lipgloss.Color("214") // Orange (topics)
	ColorMuted     = lipgloss.Color("241") // Gray
	ColorContent   = lipgloss.Color("252") // Light gray
	ColorError     = lipgloss.Color("196") // Red
	ColorBgSelect  = lipgloss.Color("236") // Dark gray background
)

// Text styles
var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary)

	Username = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	Timestamp = lipgloss.NewStyle().
			Foreground(ColorMuted)

	Content = lipgloss.NewStyle().
		Foreground(ColorContent)

	Stats = lipgloss.NewStyle().
		Foreground(ColorMuted)

	Topic = lipgloss.NewStyle().
		Foreground(ColorSecondary)

	Help = lipgloss.NewStyle().
		Foreground(ColorMuted)

	Error = lipgloss.NewStyle().
		Foreground(ColorError)

	Label = lipgloss.NewStyle().
		Foreground(ColorMuted)
)

// Layout styles
var (
	Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		Padding(0, 1)

	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Padding(1, 2)

	SelectedItem = lipgloss.NewStyle().
			Background(ColorBgSelect).
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(ColorPrimary)

	Footer = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Padding(0, 1)
)

// Spinner style
var Spinner = lipgloss.NewStyle().Foreground(ColorPrimary)

// Divider returns a horizontal divider of the given width
func Divider(width int) string {
	if width < 1 {
		width = 80
	}
	return lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(repeat("─", width))
}

// repeat is a helper to avoid importing strings just for this
func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
