package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Release represents a GitHub release
type Release struct {
	TagName    string  `json:"tag_name"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

// Asset represents a GitHub release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// TUIModel represents the main TUI model for plugin selection
type TUIModel struct {
	choices  []string
	cursor   int
	selected string
	version  string
}

func NewTUIModel(version string) TUIModel {
	files, err := ioutil.ReadDir(PluginsDir)
	if err != nil {
		log.Fatal(err)
	}

	var choices []string
	choices = append(choices, "Install a plugin")

	for _, f := range files {
		if f.Mode()&0111 != 0 { // check if executable
			choices = append(choices, f.Name())
		}
	}

	return TUIModel{
		choices: choices,
		version: version,
	}
}

// Init initializes the TUI model
func (m TUIModel) Init() tea.Cmd {
	return nil
}

// Update handles user input and updates the TUI model
func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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

				if m.selected == "" {
					return m, nil
				}

				return m.runPlugin()
			}
		}
	}
	return m, nil
}

func (m TUIModel) View() string {
	s := "Select a plugin to run:\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nPress q to quit.\n"
	s += "Use arrow keys or j/k to navigate, enter to select.\n"

	return s
}

func (m TUIModel) handlePluginInstallation() (tea.Model, tea.Cmd) {
	var url string
	if m.version == "DEVEL" {
		url = "https://api.github.com/repos/monobilisim/monokit2/releases/tags/devel"
	} else {
		url = "https://api.github.com/repos/monobilisim/monokit2/releases/latest"
	}

	client := http.Client{
		Timeout:   30 * time.Second,
		Transport: http.DefaultTransport,
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error fetching release info: %v", err)
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d", resp.StatusCode)
		return m, nil
	}

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		fmt.Printf("Error decoding release JSON: %v", err)
		return m, nil
	}

	fmt.Println(release)

	var pluginAssets []Asset
	for _, asset := range release.Assets {
		if asset.Name != "monokit2" {
			pluginAssets = append(pluginAssets, asset)
		}
	}

	for _, asset := range pluginAssets {
		fmt.Printf(asset.Name)
	}

	var choices []string
	choices = append(choices, "Cancel")

	for _, asset := range pluginAssets {
		choices = append(choices, asset.Name)
	}

	p := tea.NewProgram(TUIModel{choices: choices, version: m.version})
	m2, err := p.StartReturningModel()
	if err != nil {
		log.Fatal(err)
	}

	m3 := m2.(TUIModel)

	if m3.selected == "Cancel" || m3.selected == "" {
		return m, nil
	}

	var downloadURL string
	for _, asset := range pluginAssets {
		if asset.Name == m3.selected {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		fmt.Printf("Download URL not found for selected plugin")
		return m, nil
	}

	err = downloadAndInstallPlugin(m3.selected, downloadURL, client)
	if err != nil {
		fmt.Printf("Error installing plugin: %v", err)
		return m, nil
	}

	fmt.Printf("Plugin %s installed successfully!\n", m3.selected)
	return NewTUIModel(m.version), nil
}

func (m TUIModel) runPlugin() (tea.Model, tea.Cmd) {
	pluginPath := filepath.Join(PluginsDir, m.selected)
	cmd := exec.Command(pluginPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		log.Printf("Error running plugin %s: %v", m.selected, err)
	}
	return m, tea.Quit
}

func downloadAndInstallPlugin(pluginName, downloadURL string, client http.Client) error {
	fmt.Printf("Downloading %s...\n", pluginName)

	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("error downloading plugin: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code while downloading: %d", resp.StatusCode)
	}

	pluginPath := filepath.Join(PluginsDir, pluginName)
	outFile, err := os.Create(pluginPath)
	if err != nil {
		return fmt.Errorf("error creating plugin file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error saving plugin file: %v", err)
	}

	err = os.Chmod(pluginPath, 0755)
	if err != nil {
		return fmt.Errorf("error setting executable permission: %v", err)
	}

	return nil
}

func RunTUI(version string) error {
	model := NewTUIModel(version)
	p := tea.NewProgram(model)
	return p.Start()
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
