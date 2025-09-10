package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	lib "github.com/monobilisim/monokit2/lib"
)

type model struct {
	choices  []string
	cursor   int
	selected string
}

func initialModel() model {
	// Read plugin directory
	files, err := ioutil.ReadDir(lib.PluginsDir)
	if err != nil {
		log.Fatal(err)
	}

	var choices []string
	for _, f := range files {
		if f.Mode()&0111 != 0 { // check if executable
			choices = append(choices, f.Name())
		}
	}

	return model{choices: choices}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				pluginPath := filepath.Join(lib.PluginsDir, m.selected)
				cmd := exec.Command(pluginPath)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin
				err := cmd.Run()
				if err != nil {
					log.Printf("Error running plugin %s: %v", m.selected, err)
				}
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
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

func main() {
	lib.InitConfig()

	err := os.MkdirAll(lib.LogDir, os.ModePerm)
	if err != nil {
		panic("Failed to create log directory: " + err.Error())
	}

	if err = lib.InitializeDatabase(); err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		return
	}

	logger, err := lib.InitLogger()
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		return
	}

	logger.Info().Msg("Logger initialized successfully")

	// logger.Info().Msg("Starting the Zulip alarm worker...")
	// lib.StartZulipAlarmWorker()

	err = os.MkdirAll(lib.PluginsDir, 0755)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create plugins directory")
	}

	if len(os.Args) > 1 {
		plugin := os.Args[1]
		pluginPath := filepath.Join(lib.PluginsDir, plugin)
		if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
			fmt.Printf("Plugin %s not found\n", plugin)
		}
		if err == nil {
			cmd := exec.Command(pluginPath, os.Args[2:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			err := cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
		return
	}

	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}

	//	lib.CreateRedmineIssue(lib.Issue{
	//		ProjectId:  lib.GlobalConfig.ProjectIdentifier,
	//		TrackerId:  7,
	//		PriorityId: 5,
	//		Subject:    "monokit2 test",
	//		Description: `This is a test issue created by monokit2.
	//
	// If you see this issue, it means that the Redmine integration is working correctly.`,
	//
	//	})
}
