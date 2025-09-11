package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Release struct {
	TagName    string  `json:"tag_name"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type PluginInfo struct {
	Name             string
	InstalledVersion string
	AvailableVersion string
	HasUpdate        bool
}

type UIState int

const (
	MainMenu UIState = iota
	PluginSelection
	Installing
)

type ProgressMsg struct {
	Downloaded int64
	Total      int64
	Done       bool
	Error      error
}

type TUIModel struct {
	choices         []string
	pluginInfos     []PluginInfo
	cursor          int
	selected        string
	version         string
	state           UIState
	progress        float64
	downloadingName string
	err             error
	width           int
	height          int
	cachedRelease   *Release
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

	progressBarStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	versionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	updateAvailableStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)
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
	files, err := ioutil.ReadDir(PluginsDir)
	if err != nil {
		m.err = err
		return
	}

	var choices []string
	var pluginInfos []PluginInfo

	choices = append(choices, "Install a plugin")

	availablePlugins := make(map[string]string)
	release, err := m.fetchReleaseData()
	if err != nil {
		m.err = fmt.Errorf("GitHub API unavailable: %v (showing local plugins only)", err)
	} else if release != nil {
		releaseVersion := release.TagName
		version := strings.ToLower(m.version)
		if version == "devel" || strings.Contains(version, "devel") {
			if match := regexp.MustCompile(`[a-f0-9]{5,}`).FindString(releaseVersion); match != "" {
				releaseVersion = "devel-" + match[:5]
			} else {
				releaseVersion = "devel"
			}
		}

		for _, asset := range release.Assets {
			if asset.Name != "monokit2" {
				availablePlugins[asset.Name] = releaseVersion
			}
		}
	}

	for _, f := range files {
		if f.Mode()&0111 != 0 {
			pluginName := f.Name()
			choices = append(choices, pluginName)

			installedVersion := m.getPluginVersion(pluginName)
			availableVersion := availablePlugins[pluginName]

			hasUpdate := false
			if availableVersion != "" && installedVersion != "unknown" {
				normalizedInstalled := normalizeVersion(installedVersion)
				normalizedAvailable := normalizeVersion(availableVersion)
				if normalizedInstalled != normalizedAvailable {
					hasUpdate = true
				}
			}

			pluginInfo := PluginInfo{
				Name:             pluginName,
				InstalledVersion: installedVersion,
				AvailableVersion: availableVersion,
				HasUpdate:        hasUpdate,
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

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)

	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}

	if strings.HasPrefix(version, "devel-") {
		return version
	}

	version = strings.Split(version, " ")[0]
	return version
}

func (m *TUIModel) fetchReleaseData() (*Release, error) {
	if m.cachedRelease != nil {
		return m.cachedRelease, nil
	}

	var url string
	version := strings.ToLower(m.version)
	if version == "devel" || strings.Contains(version, "devel") {
		url = "https://api.github.com/repos/monobilisim/monokit2/releases/tags/devel"
	} else {
		url = "https://api.github.com/repos/monobilisim/monokit2/releases/latest"
	}

	client := http.Client{
		Timeout:   10 * time.Second,
		Transport: http.DefaultTransport,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return nil, fmt.Errorf("JSON decode failed: %v", err)
	}

	m.cachedRelease = &release
	return &release, nil
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
		case PluginSelection:
			return m.handlePluginSelectionInput(msg)
		case Installing:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}

	case ProgressMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.state = MainMenu
			return m, nil
		}

		if msg.Done {
			m.state = MainMenu
			m.progress = 0
			m.refreshPluginList()
			return m, nil
		}

		if msg.Total > 0 {
			m.progress = float64(msg.Downloaded) / float64(msg.Total)
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
		if m.cursor < len(m.choices) {
			m.selected = m.choices[m.cursor]

			if m.selected == "Install a plugin" {
				return m.handlePluginInstallation()
			}

			if m.selected != "" {
				return m.runPlugin()
			}
		}
	case "r":
		m.cachedRelease = nil
		m.refreshPluginList()
	}
	return m, nil
}

func (m TUIModel) handlePluginSelectionInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.state = MainMenu
		m.cursor = 0
		m.refreshPluginList()
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.choices)-1 {
			m.cursor++
		}
	case "enter", " ":
		if m.cursor < len(m.choices) {
			m.selected = m.choices[m.cursor]

			if m.selected == "Cancel" {
				m.state = MainMenu
				m.cursor = 0
				m.refreshPluginList()
				return m, nil
			}

			if m.selected != "" {
				return m.installSelectedPlugin()
			}
		}
	}
	return m, nil
}

func (m TUIModel) View() string {
	switch m.state {
	case MainMenu:
		return m.renderMainMenu()
	case PluginSelection:
		return m.renderPluginSelection()
	case Installing:
		return m.renderInstallProgress()
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

	s.WriteString("Select a plugin to run:\n\n")

	for i, choice := range m.choices {
		cursor := "  "
		style := normalStyle

		if m.cursor == i {
			cursor = "> "
			style = selectedStyle
		}

		if choice == "Install a plugin" {
			s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(choice)))
		} else {
			var pluginInfo *PluginInfo
			for j := range m.pluginInfos {
				if m.pluginInfos[j].Name == choice {
					pluginInfo = &m.pluginInfos[j]
					break
				}
			}

			versionText := ""
			if pluginInfo != nil && pluginInfo.InstalledVersion != "unknown" {
				displayVersion := pluginInfo.InstalledVersion
				if !strings.HasPrefix(displayVersion, "v") && !strings.HasPrefix(displayVersion, "devel") {
					displayVersion = "v" + displayVersion
				}
				versionText = versionStyle.Render(fmt.Sprintf(" (%s)", displayVersion))
				if pluginInfo.HasUpdate && pluginInfo.AvailableVersion != "" {
					versionText += updateAvailableStyle.Render(" [update available]")
				}
			}

			s.WriteString(fmt.Sprintf("%s%s%s\n", cursor, style.Render(choice), versionText))
		}
	}

	s.WriteString("\n")
	s.WriteString("Press q to quit, r to refresh, enter to select.\n")
	s.WriteString("Use arrow keys or j/k to navigate.\n")

	return s.String()
}

func (m TUIModel) renderPluginSelection() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("📦 Available Plugins"))
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		s.WriteString("\n\n")
	}

	s.WriteString("Select a plugin to install:\n\n")

	for i, choice := range m.choices {
		cursor := "  "
		style := normalStyle

		if m.cursor == i {
			cursor = "> "
			style = selectedStyle
		}

		s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(choice)))
	}

	s.WriteString("\n")
	s.WriteString("Press esc to go back, enter to select.\n")
	s.WriteString("Use arrow keys or j/k to navigate.\n")

	return s.String()
}

func (m TUIModel) renderInstallProgress() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("⬇️  Installing Plugin"))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("Installing: %s\n\n", m.downloadingName))

	progressWidth := 50
	if m.width > 0 && m.width-20 < progressWidth {
		progressWidth = m.width - 20
	}

	filled := int(m.progress * float64(progressWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", progressWidth-filled)

	progressText := fmt.Sprintf("Progress: %.1f%%", m.progress*100)

	s.WriteString(progressBarStyle.Render(bar))
	s.WriteString("\n")
	s.WriteString(progressText)
	s.WriteString("\n\n")

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		s.WriteString("\n")
	}

	s.WriteString("Press Ctrl+C to cancel.\n")

	return s.String()
}

func (m TUIModel) handlePluginInstallation() (tea.Model, tea.Cmd) {
	release, err := m.fetchReleaseData()
	if err != nil {
		m.err = fmt.Errorf("error fetching release info: %v", err)
		return m, nil
	}
	if release == nil {
		m.err = fmt.Errorf("error fetching release info: release is nil")
		return m, nil
	}

	var pluginAssets []Asset
	for _, asset := range release.Assets {
		if asset.Name != "monokit2" {
			pluginAssets = append(pluginAssets, asset)
		}
	}

	var choices []string
	choices = append(choices, "Cancel")

	for _, asset := range pluginAssets {
		installedVersion := m.getPluginVersion(asset.Name)
		versionText := ""
		if installedVersion != "unknown" {
			availableVersion := release.TagName
			version := strings.ToLower(m.version)
			if version == "devel" || strings.Contains(version, "devel") {
				if match := regexp.MustCompile(`[a-f0-9]{5,}`).FindString(availableVersion); match != "" {
					availableVersion = "devel-" + match[:5]
				} else {
					availableVersion = "devel"
				}
			}

			normalizedInstalled := normalizeVersion(installedVersion)
			normalizedAvailable := normalizeVersion(availableVersion)

			if normalizedInstalled != normalizedAvailable {
				versionText = fmt.Sprintf(" (installed: %s → available: %s)", installedVersion, availableVersion)
			} else {
				versionText = fmt.Sprintf(" (current: %s)", installedVersion)
			}
		}

		choices = append(choices, asset.Name+versionText)
	}

	m.state = PluginSelection
	m.choices = choices
	m.cursor = 0
	m.err = nil

	return m, nil
}

func (m TUIModel) installSelectedPlugin() (tea.Model, tea.Cmd) {
	pluginName := strings.Split(m.selected, " (")[0]

	release, err := m.fetchReleaseData()
	if err != nil {
		m.err = fmt.Errorf("error fetching release info: %v", err)
		return m, nil
	}
	if release == nil {
		m.err = fmt.Errorf("error fetching release info: release is nil")
		return m, nil
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == pluginName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		m.err = fmt.Errorf("download URL not found for selected plugin")
		return m, nil
	}

	m.state = Installing
	m.downloadingName = pluginName
	m.progress = 0
	m.err = nil

	return m, m.downloadWithProgress(pluginName, downloadURL)
}

func (m TUIModel) downloadWithProgress(pluginName, downloadURL string) tea.Cmd {
	return func() tea.Msg {
		client := http.Client{
			Timeout:   5 * time.Minute,
			Transport: http.DefaultTransport,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
		if err != nil {
			return ProgressMsg{Error: fmt.Errorf("error creating request: %v", err)}
		}

		resp, err := client.Do(req)
		if err != nil {
			return ProgressMsg{Error: fmt.Errorf("error downloading plugin: %v", err)}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return ProgressMsg{Error: fmt.Errorf("unexpected status code: %d", resp.StatusCode)}
		}

		pluginPath := filepath.Join(PluginsDir, pluginName)
		outFile, err := os.Create(pluginPath)
		if err != nil {
			return ProgressMsg{Error: fmt.Errorf("error creating plugin file: %v", err)}
		}
		defer outFile.Close()

		totalSize := resp.ContentLength
		var downloaded int64

		buffer := make([]byte, 32*1024)
	downloadLoop:
		for {
			select {
			case <-ctx.Done():
				return ProgressMsg{Error: fmt.Errorf("download cancelled")}
			default:
				n, err := resp.Body.Read(buffer)
				if n > 0 {
					_, writeErr := outFile.Write(buffer[:n])
					if writeErr != nil {
						return ProgressMsg{Error: fmt.Errorf("error writing to file: %v", writeErr)}
					}
					downloaded += int64(n)
				}

				if err != nil {
					if err == io.EOF {
						break downloadLoop
					}
					return ProgressMsg{Error: fmt.Errorf("error reading response: %v", err)}
				}
			}
		}

		err = os.Chmod(pluginPath, 0755)
		if err != nil {
			return ProgressMsg{Error: fmt.Errorf("error setting executable permission: %v", err)}
		}

		return ProgressMsg{
			Downloaded: downloaded,
			Total:      totalSize,
			Done:       true,
		}
	}
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
