package initializer

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type stepStatus int

const (
	statusPending stepStatus = iota
	statusRunning
	statusDone
	statusFailed
)

type step struct {
	name   string
	action func() error
}

type stepDoneMsg struct {
	index int
	err   error
}

type model struct {
	state    *state
	steps    []step
	statuses []stepStatus
	current  int
	finished bool
	err      error
}

func newModel(st *state) *model {
	steps := []step{
		{name: "Clone template repository", action: func() error { return cloneTemplate(st.ProjectDir) }},
		{name: "Remove template .git metadata", action: func() error { return removeTemplateGitMetadata(st.ProjectDir) }},
		{name: "Rewrite module path", action: func() error { return rewriteModulePath(st.ProjectDir, st.ModulePath) }},
		{name: "Initialize new Git repository (master)", action: func() error { return reinitializeGitRepository(st.ProjectDir) }},
		{name: "Run go mod tidy", action: func() error { return tidyGoModule(st.ProjectDir) }},
	}

	statuses := make([]stepStatus, len(steps))
	if len(statuses) > 0 {
		statuses[0] = statusRunning
	}

	return &model{
		state:    st,
		steps:    steps,
		statuses: statuses,
	}
}

func (m *model) Init() tea.Cmd {
	if len(m.steps) == 0 {
		m.finished = true
		return tea.Quit
	}
	return runStepCmd(m.steps[0], 0)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if !m.finished {
				m.err = errors.New("initialization cancelled")
				if m.current >= 0 && m.current < len(m.statuses) && m.statuses[m.current] == statusRunning {
					m.statuses[m.current] = statusFailed
				}
			}
			m.finished = true
			return m, tea.Quit
		}
	case stepDoneMsg:
		if msg.err != nil {
			m.statuses[msg.index] = statusFailed
			m.err = msg.err
			m.finished = true
			return m, tea.Quit
		}

		m.statuses[msg.index] = statusDone
		if msg.index == len(m.steps)-1 {
			m.finished = true
			return m, tea.Quit
		}

		m.current = msg.index + 1
		m.statuses[m.current] = statusRunning
		return m, runStepCmd(m.steps[m.current], m.current)
	}

	return m, nil
}

func (m *model) View() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Bamboo initializing: %s\n\n", m.state.ModulePath))

	for index, item := range m.steps {
		icon := "[ ]"
		switch m.statuses[index] {
		case statusRunning:
			icon = "[>]"
		case statusDone:
			icon = "[x]"
		case statusFailed:
			icon = "[!]"
		}
		builder.WriteString(fmt.Sprintf(" %s %s\n", icon, item.name))
	}

	if m.err != nil {
		builder.WriteString(fmt.Sprintf("\nFailed: %v\n", m.err))
		return builder.String()
	}

	if m.finished {
		builder.WriteString(fmt.Sprintf("\nDone.\nProject directory: %s\n", m.state.ProjectDir))
		return builder.String()
	}

	builder.WriteString("\nPress Ctrl+C to cancel.\n")
	return builder.String()
}

func runStepCmd(currentStep step, index int) tea.Cmd {
	return func() tea.Msg {
		return stepDoneMsg{
			index: index,
			err:   currentStep.action(),
		}
	}
}
