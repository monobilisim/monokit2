package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	ColorPrimary   = lipgloss.Color("#7D56F4")
	ColorSecondary = lipgloss.Color("#00F5D4")
	ColorDark      = lipgloss.Color("#1A1A24")
	ColorGray      = lipgloss.Color("#454141")
	ColorSuccess   = lipgloss.Color("#04B575")
	ColorWarning   = lipgloss.Color("#FFB84C")
	ColorError     = lipgloss.Color("#FF4D4D")
	ColorInfo      = lipgloss.Color("#00A8FF")
)

var (
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2).
			MarginBottom(1)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(ColorPrimary).
			Padding(0, 1).
			Bold(true)

	SuccessBadge = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true).SetString("✔ SUCCESS ")
	ErrorBadge   = lipgloss.NewStyle().Foreground(ColorError).Bold(true).SetString("✖ ERROR   ")
	WarningBadge = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true).SetString("⚠ WARNING ")
	InfoBadge    = lipgloss.NewStyle().Foreground(ColorInfo).Bold(true).SetString("ℹ INFO    ")
)

type KV struct {
	Key   string
	Value string
}

func RenderPluginCard(pluginName string, content string) string {
	header := HeaderStyle.Render(strings.ToUpper(pluginName))

	if content == "" {
		content = lipgloss.NewStyle().Foreground(ColorGray).Render("No output.")
	}

	fullLayout := fmt.Sprintf("%s\n\n%s", header, content)
	return CardStyle.Render(fullLayout)
}

func Log(badge lipgloss.Style, message string) string {
	return fmt.Sprintf("%s %s", badge.String(), message)
}

func RenderTable(headers []string, rows [][]string) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(ColorDark).
		Background(ColorPrimary).
		Bold(true).
		Align(lipgloss.Center)

	cellStyle := lipgloss.NewStyle().Padding(0, 1)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorGray)).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			return cellStyle
		}).
		Headers(headers...).
		Rows(rows...)

	return t.Render()
}

func RenderKeyValueList(items []KV) string {
	var sb strings.Builder

	keyStyle := lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	colonStyle := lipgloss.NewStyle().Foreground(ColorGray).SetString(" : ")

	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))

	maxLen := 0
	for _, item := range items {
		if len(item.Key) > maxLen {
			maxLen = len(item.Key)
		}
	}

	for _, item := range items {
		paddedKey := keyStyle.Copy().Width(maxLen).Render(item.Key)
		line := fmt.Sprintf("%s%s%s\n", paddedKey, colonStyle.String(), valueStyle.Render(item.Value))
		sb.WriteString(line)
	}
	return sb.String()
}
