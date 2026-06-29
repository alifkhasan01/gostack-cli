package printer

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1")).Bold(true)
	blue   = lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa")).Bold(true)
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af")).Bold(true)
	dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086"))
)

// Step prints a step label with an emoji.
func Step(emoji, msg string) {
	fmt.Printf("\n%s  %s\n", emoji, blue.Render(msg))
}

// Success prints a green success message.
func Success(msg string) {
	fmt.Printf("\n%s  %s\n", "✅", green.Render(msg))
}

// Warn prints a yellow warning.
func Warn(msg string) {
	fmt.Printf("%s  %s\n", "⚠️ ", yellow.Render(msg))
}

// Dim prints a dimmed info line.
func Dim(msg string) {
	fmt.Println(dim.Render("  " + msg))
}

// Summary prints the post-create summary box.
func Summary(projectName, dir string) {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#89b4fa")).
		Padding(1, 3)

	content := fmt.Sprintf(
		"%s\n\n%s\n  %s\n  %s\n\n%s\n  %s",
		green.Render("🎉 Project created successfully!"),
		blue.Render("Next steps:"),
		dim.Render("$ ")+fmt.Sprintf("cd %s", projectName),
		dim.Render("$ ")+"go run ./cmd/api",
		blue.Render("Project location:"),
		dir,
	)

	fmt.Println(box.Render(content))
}
