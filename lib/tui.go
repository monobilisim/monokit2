package lib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PluginInfo struct {
	Name    string
	Version string
}

type UIState int

const (
	MainMenu UIState = iota
	Running
)

type TUIModel struct {
	choices     []string
	pluginInfos []PluginInfo
	cursor      int
	selected    string
	version     string
	state       UIState
	err         error
	width       int
	height      int
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	versionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF")).
			Italic(true)
)

func NewTUIModel(version string) TUIModel {
	model := TUIModel{
		version: version,
		state:   MainMenu,
	}
	model.refreshPluginList()
	return model
}

func (m *TUIModel) refreshPluginList() {
	m.err = nil

	if _, err := os.Stat(PluginsDir); os.IsNotExist(err) {
		err := os.MkdirAll(PluginsDir, 0755)
		if err != nil {
			m.err = fmt.Errorf("failed to create plugins directory: %v", err)
			return
		}
	}

	entries, err := os.ReadDir(PluginsDir)
	if err != nil {
		m.err = err
		return
	}

	var choices []string
	var pluginInfos []PluginInfo

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.Mode()&0111 != 0 {
			pluginName := entry.Name()
			choices = append(choices, pluginName)

			version := m.getPluginVersion(pluginName)
			pluginInfo := PluginInfo{
				Name:    pluginName,
				Version: version,
			}
			pluginInfos = append(pluginInfos, pluginInfo)
		}
	}

	m.choices = choices
	m.pluginInfos = pluginInfos
}

func (m *TUIModel) getPluginVersion(pluginName string) string {
	pluginPath := filepath.Join(PluginsDir, pluginName)
	cmd := exec.Command(pluginPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	version := strings.TrimSpace(string(output))

	if strings.Contains(version, "devel") || len(version) > 10 {
		if match := regexp.MustCompile(`[a-f0-9]{5,}`).FindString(version); match != "" {
			return "devel-" + match[:5]
		}
		return "devel"
	}

	return version
}

func (m TUIModel) Init() tea.Cmd {
	return nil
}

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch m.state {
		case MainMenu:
			return m.handleMainMenuInput(msg)
		case Running:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m TUIModel) handleMainMenuInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.choices)-1 {
			m.cursor++
		}
	case "enter", " ":
		if m.cursor < len(m.choices) && len(m.choices) > 0 {
			m.selected = m.choices[m.cursor]
			if m.selected != "" {
				return m.runPlugin()
			}
		}
	case "r":
		m.refreshPluginList()
	}
	return m, nil
}

func (m TUIModel) View() string {
	switch m.state {
	case MainMenu:
		return m.renderMainMenu()
	case Running:
		return m.renderRunning()
	}
	return ""
}

func (m TUIModel) renderMainMenu() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("🔧 Monokit2 Plugin Manager"))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		s.WriteString("\n\n")
	}

	if len(m.choices) == 0 {
		s.WriteString(infoStyle.Render("No plugins installed."))
		s.WriteString("\n\n")
		s.WriteString(fmt.Sprintf("Place executable plugins in: %s\n", PluginsDir))
	} else {
		s.WriteString("Select a plugin to run:\n\n")

		for i, choice := range m.choices {
			cursor := "  "
			style := normalStyle

			if m.cursor == i {
				cursor = "> "
				style = selectedStyle
			}

			var pluginInfo *PluginInfo
			for j := range m.pluginInfos {
				if m.pluginInfos[j].Name == choice {
					pluginInfo = &m.pluginInfos[j]
					break
				}
			}

			versionText := ""
			if pluginInfo != nil && pluginInfo.Version != "unknown" {
				displayVersion := pluginInfo.Version
				if !strings.HasPrefix(displayVersion, "v") && !strings.HasPrefix(displayVersion, "devel") {
					displayVersion = "v" + displayVersion
				}
				versionText = versionStyle.Render(fmt.Sprintf(" (%s)", displayVersion))
			}

			s.WriteString(fmt.Sprintf("%s%s%s\n", cursor, style.Render(choice), versionText))
		}
	}

	s.WriteString("\n")
	s.WriteString("Press q to quit, r to refresh, enter to select.\n")
	s.WriteString("Use arrow keys or j/k to navigate.\n")

	return s.String()
}

func (m TUIModel) renderRunning() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("▶️  Running Plugin"))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("Running: %s\n", m.selected))
	s.WriteString("\nPress Ctrl+C to cancel.\n")

	return s.String()
}

func (m TUIModel) runPlugin() (tea.Model, tea.Cmd) {
	pluginPath := filepath.Join(PluginsDir, m.selected)
	cmd := exec.Command(pluginPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		m.err = fmt.Errorf("error running plugin %s: %v", m.selected, err)
		return m, nil
	}
	return m, tea.Quit
}

func RunTUI(version string) error {
	model := NewTUIModel(version)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func RunPlugin(pluginName string, args []string) error {
	pluginPath := filepath.Join(PluginsDir, pluginName)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return fmt.Errorf("plugin %s not found", pluginName)
	}

	cmd := exec.Command(pluginPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func ListPlugins() ([]PluginInfo, error) {
	if _, err := os.Stat(PluginsDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(PluginsDir)
	if err != nil {
		return nil, err
	}

	var plugins []PluginInfo

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.Mode()&0111 != 0 {
			pluginName := entry.Name()
			pluginPath := filepath.Join(PluginsDir, pluginName)

			version := "unknown"
			cmd := exec.Command(pluginPath, "version")
			output, err := cmd.Output()
			if err == nil {
				version = strings.TrimSpace(string(output))
				if strings.Contains(version, "devel") || len(version) > 10 {
					if match := regexp.MustCompile(`[a-f0-9]{5,}`).FindString(version); match != "" {
						version = "devel-" + match[:5]
					} else {
						version = "devel"
					}
				}
			}

			plugins = append(plugins, PluginInfo{
				Name:    pluginName,
				Version: version,
			})
		}
	}

	return plugins, nil
}
