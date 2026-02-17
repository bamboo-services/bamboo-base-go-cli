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
		return errors.New("内部错误：TUI 最终模型类型不匹配")
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
		return nil, fmt.Errorf("无效的包名：%q", modulePath)
	}

	projectDir := filepath.Join(workDir, projectName)
	return &state{
		ModulePath: modulePath,
		ProjectDir: projectDir,
	}, nil
}

func validateModulePath(modulePath string) error {
	if modulePath == "" {
		return errors.New("包名不能为空")
	}
	if strings.Contains(modulePath, " ") {
		return fmt.Errorf("包名不能包含空格：%q", modulePath)
	}
	if strings.HasPrefix(modulePath, "/") || strings.HasSuffix(modulePath, "/") {
		return fmt.Errorf("包名不能以 '/' 开头或结尾：%q", modulePath)
	}
	parts := strings.Split(modulePath, "/")
	if len(parts) < 2 {
		return fmt.Errorf("包名格式应为 host/path，当前为：%q", modulePath)
	}
	if !strings.Contains(parts[0], ".") {
		return fmt.Errorf("包名第一段应包含主机名，当前为：%q", modulePath)
	}
	for _, part := range parts {
		if part == "" {
			return fmt.Errorf("包名包含空路径段：%q", modulePath)
		}
	}

	return nil
}
