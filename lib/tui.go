package lib

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	githubOwner = "monobilisim"
	githubRepo  = "monokit2"
)

type PluginInfo struct {
	Name        string
	Version     string
	DisplayName string
}

type RemotePlugin struct {
	Name        string
	DownloadURL string
	FileName    string
}

type UIState int

const (
	StateMainMenu UIState = iota
	StatePluginSelection
	StateInstalling
	StateRunning
)

type ProgressMsg struct {
	Downloaded int64
	Total      int64
	Done       bool
	Error      error
}

type FetchPluginsMsg struct {
	Plugins []RemotePlugin
	Error   error
}

type TUIModel struct {
	choices         []string
	pluginInfos     []PluginInfo
	remotePlugins   []RemotePlugin
	cursor          int
	selected        string
	version         string
	state           UIState
	progress        float64
	downloadingName string
	err             error
	width           int
	height          int
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

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF")).
			Italic(true)

	updateAvailableStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF00")).
				Bold(true)
)

func NewTUIModel(version string) TUIModel {
	model := TUIModel{
		version: version,
		state:   StateMainMenu,
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

	// Add "Install a plugin" option first
	choices = append(choices, "Install a plugin")

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

			// Skip monokit2 itself if present
			if pluginName == "monokit2" {
				continue
			}

			choices = append(choices, pluginName)

			version := m.getPluginVersion(pluginName)
			pluginInfo := PluginInfo{
				Name:        pluginName,
				Version:     version,
				DisplayName: pluginName, // Now plugins are saved with base name
			}
			pluginInfos = append(pluginInfos, pluginInfo)
		}
	}

	m.choices = choices
	m.pluginInfos = pluginInfos
}

// extractPluginDisplayName extracts the base plugin name from full filename
// e.g., "osHealth_devel_linux_amd64" -> "osHealth"
func extractPluginDisplayName(pluginName string) string {
	// Remove .exe extension if present
	cleanName := strings.TrimSuffix(pluginName, ".exe")

	// Split by underscore and take the first part
	parts := strings.Split(cleanName, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return cleanName
}

// isPluginForCurrentPlatform checks if a plugin matches the current OS and architecture.
// Plugin name format: name_version_os_arch (e.g., osHealth_devel_linux_amd64)
// Also handles .exe extension for Windows binaries.
func isPluginForCurrentPlatform(pluginName, currentOS, currentArch string) bool {
	// Remove .exe extension if present (for Windows binaries)
	cleanName := pluginName
	if strings.HasSuffix(cleanName, ".exe") {
		cleanName = strings.TrimSuffix(cleanName, ".exe")
	}

	// Expected suffix format: _os_arch
	expectedSuffix := fmt.Sprintf("_%s_%s", currentOS, currentArch)

	// If the plugin name ends with current platform suffix, it's a match
	if strings.HasSuffix(cleanName, expectedSuffix) {
		return true
	}

	// Check if the plugin has any known platform suffix (meaning it's for a different platform)
	knownOSes := []string{"linux", "darwin", "windows", "freebsd", "openbsd", "netbsd"}
	knownArchs := []string{"amd64", "arm64", "386", "arm", "ppc64le", "s390x", "riscv64"}

	for _, osName := range knownOSes {
		for _, arch := range knownArchs {
			suffix := fmt.Sprintf("_%s_%s", osName, arch)
			if strings.HasSuffix(cleanName, suffix) {
				// Plugin has a platform suffix but it's not for current platform
				return false
			}
		}
	}

	// Plugin has no platform suffix - consider it a generic/compatible plugin
	return true
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

// fetchLatestReleaseTag scrapes GitHub releases page to find the latest release tag
func fetchLatestReleaseTag(useDevel bool) (string, error) {
	if useDevel {
		return "devel", nil
	}

	url := fmt.Sprintf("https://github.com/%s/%s/releases/latest", githubOwner, githubRepo)

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects, we want to capture the redirect URL
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %v", err)
	}
	defer resp.Body.Close()

	// GitHub redirects /releases/latest to /releases/tag/{tag}
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently {
		location := resp.Header.Get("Location")
		// Extract tag from URL like: https://github.com/owner/repo/releases/tag/v1.0.0
		parts := strings.Split(location, "/tag/")
		if len(parts) == 2 {
			return parts[1], nil
		}
	}

	// If no redirect, try to parse the page
	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response: %v", err)
		}

		// Look for tag in the HTML
		tagRegex := regexp.MustCompile(`/releases/tag/([^"]+)`)
		matches := tagRegex.FindSubmatch(body)
		if len(matches) >= 2 {
			return string(matches[1]), nil
		}
	}

	return "", fmt.Errorf("could not determine latest release tag")
}

// fetchAvailablePlugins scrapes GitHub releases page to get available plugin assets
func fetchAvailablePlugins(tag string) ([]RemotePlugin, error) {
	url := fmt.Sprintf("https://github.com/%s/%s/releases/expanded_assets/%s", githubOwner, githubRepo, tag)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release assets: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch release assets: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse asset links from HTML
	// Looking for links like: href="/monobilisim/monokit2/releases/download/devel/osHealth_devel_linux_amd64"
	assetRegex := regexp.MustCompile(`href="(/` + githubOwner + `/` + githubRepo + `/releases/download/[^"]+)"`)
	matches := assetRegex.FindAllSubmatch(body, -1)

	currentOS := runtime.GOOS
	currentArch := runtime.GOARCH

	var plugins []RemotePlugin
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) >= 2 {
			path := string(match[1])
			fileName := filepath.Base(path)

			// Skip if already seen
			if seen[fileName] {
				continue
			}
			seen[fileName] = true

			// Skip monokit2 itself (the main binary)
			baseName := extractPluginDisplayName(fileName)
			if baseName == "monokit2" {
				continue
			}

			// Only include plugins for current platform
			if !isPluginForCurrentPlatform(fileName, currentOS, currentArch) {
				continue
			}

			downloadURL := fmt.Sprintf("https://github.com%s", path)
			plugins = append(plugins, RemotePlugin{
				Name:        baseName,
				DownloadURL: downloadURL,
				FileName:    fileName,
			})
		}
	}

	return plugins, nil
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
		case StateMainMenu:
			return m.handleMainMenuInput(msg)
		case StatePluginSelection:
			return m.handlePluginSelectionInput(msg)
		case StateInstalling:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		case StateRunning:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}

	case FetchPluginsMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.state = StateMainMenu
			return m, nil
		}
		m.remotePlugins = msg.Plugins
		m.buildPluginSelectionChoices()
		m.state = StatePluginSelection
		m.cursor = 0
		return m, nil

	case ProgressMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.state = StateMainMenu
			return m, nil
		}

		if msg.Done {
			m.state = StateMainMenu
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

func (m *TUIModel) buildPluginSelectionChoices() {
	var choices []string
	choices = append(choices, "← Back")

	for _, plugin := range m.remotePlugins {
		// Check if already installed
		installedVersion := ""
		for _, installed := range m.pluginInfos {
			if installed.DisplayName == plugin.Name {
				installedVersion = installed.Version
				break
			}
		}

		displayText := plugin.Name
		if installedVersion != "" {
			displayText = fmt.Sprintf("%s (installed: %s)", plugin.Name, installedVersion)
		}
		choices = append(choices, displayText)
	}

	m.choices = choices
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

			if m.selected == "Install a plugin" {
				return m.handlePluginInstallation()
			}

			if m.selected != "" {
				return m.runPlugin()
			}
		}
	case "r":
		m.refreshPluginList()
	}
	return m, nil
}

func (m TUIModel) handlePluginSelectionInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.state = StateMainMenu
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

			if m.selected == "← Back" {
				m.state = StateMainMenu
				m.cursor = 0
				m.refreshPluginList()
				return m, nil
			}

			// Find the selected plugin
			selectedName := strings.Split(m.selected, " (")[0]
			for _, plugin := range m.remotePlugins {
				if plugin.Name == selectedName {
					return m.installPlugin(plugin)
				}
			}
		}
	}
	return m, nil
}

func (m TUIModel) View() string {
	switch m.state {
	case StateMainMenu:
		return m.renderMainMenu()
	case StatePluginSelection:
		return m.renderPluginSelection()
	case StateInstalling:
		return m.renderInstallProgress()
	case StateRunning:
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

	s.WriteString("Select an option:\n\n")

	for i, choice := range m.choices {
		cursor := "  "
		style := normalStyle

		if m.cursor == i {
			cursor = "> "
			style = selectedStyle
		}

		if choice == "Install a plugin" {
			s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render("📦 "+choice)))
		} else {
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

			displayName := choice
			if pluginInfo != nil && pluginInfo.DisplayName != "" {
				displayName = pluginInfo.DisplayName
			}

			s.WriteString(fmt.Sprintf("%s%s%s\n", cursor, style.Render(displayName), versionText))
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

		if choice == "← Back" {
			s.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(choice)))
		} else {
			// Check if it's installed
			if strings.Contains(choice, "(installed:") {
				s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, style.Render(strings.Split(choice, " (")[0]), versionStyle.Render("("+strings.Split(choice, "(")[1])))
			} else {
				s.WriteString(fmt.Sprintf("%s%s %s\n", cursor, style.Render(choice), updateAvailableStyle.Render("[new]")))
			}
		}
	}

	s.WriteString("\n")
	s.WriteString("Press esc to go back, enter to install.\n")
	s.WriteString("Use arrow keys or j/k to navigate.\n")

	return s.String()
}

func (m TUIModel) renderInstallProgress() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("⬇️  Installing Plugin"))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("Downloading: %s\n\n", m.downloadingName))

	progressWidth := 50
	if m.width > 0 && m.width-20 < progressWidth {
		progressWidth = m.width - 20
	}
	if progressWidth < 10 {
		progressWidth = 10
	}

	filled := int(m.progress * float64(progressWidth))
	if filled > progressWidth {
		filled = progressWidth
	}
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

func (m TUIModel) renderRunning() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("▶️  Running Plugin"))
	s.WriteString("\n\n")

	s.WriteString(fmt.Sprintf("Running: %s\n", m.selected))
	s.WriteString("\nPress Ctrl+C to cancel.\n")

	return s.String()
}

func (m TUIModel) handlePluginInstallation() (tea.Model, tea.Cmd) {
	m.err = nil

	// Determine if we should use devel or latest release
	useDevel := strings.Contains(strings.ToLower(m.version), "devel")

	return m, func() tea.Msg {
		tag, err := fetchLatestReleaseTag(useDevel)
		if err != nil {
			return FetchPluginsMsg{Error: err}
		}

		plugins, err := fetchAvailablePlugins(tag)
		if err != nil {
			return FetchPluginsMsg{Error: err}
		}

		if len(plugins) == 0 {
			return FetchPluginsMsg{Error: fmt.Errorf("no plugins found for %s/%s", runtime.GOOS, runtime.GOARCH)}
		}

		return FetchPluginsMsg{Plugins: plugins}
	}
}

func (m TUIModel) installPlugin(plugin RemotePlugin) (tea.Model, tea.Cmd) {
	m.state = StateInstalling
	m.downloadingName = plugin.Name
	m.progress = 0
	m.err = nil

	return m, m.downloadWithProgress(plugin)
}

func (m TUIModel) downloadWithProgress(plugin RemotePlugin) tea.Cmd {
	return func() tea.Msg {
		client := http.Client{
			Timeout: 5 * time.Minute,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", plugin.DownloadURL, nil)
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

		// Save with base name only (e.g., "osHealth" instead of "osHealth_devel_linux_amd64")
		pluginPath := filepath.Join(PluginsDir, plugin.Name)
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

			// Skip monokit2 itself if present
			if pluginName == "monokit2" {
				continue
			}

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
				Name:        pluginName,
				Version:     version,
				DisplayName: pluginName,
			})
		}
	}

	return plugins, nil
}
