package initializer

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	templateRepoURL = "https://github.com/bamboo-services/bamboo-base-go-template"
	templateModule  = "github.com/bamboo-services/bamboo-base-go-template"
)

type state struct {
	ModulePath string
	ProjectDir string
}

func Run(modulePath string, workDir string) error {
	st, err := newState(modulePath, workDir)
	if err != nil {
		return err
	}

	program := tea.NewProgram(newModel(st), tea.WithInput(nil))
	finalModel, err := program.Run()
	if err != nil {
		return err
	}

	m, ok := finalModel.(*model)
	if !ok {
		return errors.New("internal error: unexpected final tui model type")
	}

	if m.err != nil {
		return m.err
	}

	return nil
}

func newState(modulePath string, workDir string) (*state, error) {
	modulePath = strings.TrimSpace(modulePath)
	if err := validateModulePath(modulePath); err != nil {
		return nil, err
	}

	projectName := path.Base(modulePath)
	projectName = strings.TrimSuffix(projectName, ".git")
	if projectName == "." || projectName == "/" || projectName == "" {
		return nil, fmt.Errorf("invalid package name: %q", modulePath)
	}

	projectDir := filepath.Join(workDir, projectName)
	return &state{
		ModulePath: modulePath,
		ProjectDir: projectDir,
	}, nil
}

func validateModulePath(modulePath string) error {
	if modulePath == "" {
		return errors.New("package name is required")
	}
	if strings.Contains(modulePath, " ") {
		return fmt.Errorf("package name must not include spaces: %q", modulePath)
	}
	if strings.HasPrefix(modulePath, "/") || strings.HasSuffix(modulePath, "/") {
		return fmt.Errorf("package name must not start or end with '/': %q", modulePath)
	}
	parts := strings.Split(modulePath, "/")
	if len(parts) < 2 {
		return fmt.Errorf("package name must look like host/path, got: %q", modulePath)
	}
	if !strings.Contains(parts[0], ".") {
		return fmt.Errorf("package name must include a host in the first segment, got: %q", modulePath)
	}
	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("package name contains an empty path segment: %q", modulePath)
		}
	}

	return nil
}
