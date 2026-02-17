package initializer

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	gocmd "github.com/go-cmd/cmd"
)

func cloneTemplate(projectDir string) error {
	if _, err := os.Stat(projectDir); err == nil {
		return fmt.Errorf("目标目录已存在：%s", projectDir)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("检查目标目录失败：%w", err)
	}

	if err := runCommand("", "git", "clone", "--depth=1", templateRepoURL, projectDir); err != nil {
		return fmt.Errorf("克隆模板仓库失败：%w", err)
	}

	return nil
}

func removeTemplateGitMetadata(projectDir string) error {
	gitDir := filepath.Join(projectDir, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return fmt.Errorf("删除模板 .git 目录失败：%w", err)
	}
	return nil
}

func rewriteModulePath(projectDir string, modulePath string) error {
	return filepath.WalkDir(projectDir, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !shouldRewriteFile(filePath) {
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("读取文件失败（%s）：%w", filePath, err)
		}

		updated := bytes.ReplaceAll(content, []byte(templateModule), []byte(modulePath))
		if bytes.Equal(content, updated) {
			return nil
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("获取文件信息失败（%s）：%w", filePath, err)
		}

		if err := os.WriteFile(filePath, updated, fileInfo.Mode().Perm()); err != nil {
			return fmt.Errorf("写入文件失败（%s）：%w", filePath, err)
		}

		return nil
	})
}

func reinitializeGitRepository(projectDir string) error {
	if err := runCommand(projectDir, "git", "init", "-b", "master"); err == nil {
		return nil
	}

	if err := runCommand(projectDir, "git", "init"); err != nil {
		return fmt.Errorf("初始化 git 仓库失败：%w", err)
	}

	if err := runCommand(projectDir, "git", "symbolic-ref", "HEAD", "refs/heads/master"); err != nil {
		if renameErr := runCommand(projectDir, "git", "branch", "-M", "master"); renameErr != nil {
			return fmt.Errorf("设置 git 默认分支为 master 失败：%w", err)
		}
	}

	return nil
}

func tidyGoModule(projectDir string) error {
	if err := runCommand(projectDir, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("执行 go mod tidy 失败：%w", err)
	}
	return nil
}

func shouldRewriteFile(filePath string) bool {
	baseName := filepath.Base(filePath)
	if baseName == "go.mod" || baseName == "go.sum" || baseName == "README.md" || baseName == "Makefile" {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go", ".mod", ".sum", ".md", ".txt", ".yaml", ".yml", ".toml", ".env":
		return true
	default:
		return false
	}
}

func runCommand(dir string, name string, args ...string) error {
	timeout := 3 * time.Minute
	options := gocmd.Options{
		Buffered:  true,
		Streaming: false,
	}
	if dir != "" {
		options.BeforeExec = []func(command *exec.Cmd){
			func(command *exec.Cmd) {
				command.Dir = dir
			},
		}
	}

	cmd := gocmd.NewCmdOptions(options, name, args...)
	statusChan := cmd.Start()

	select {
	case status := <-statusChan:
		if status.Error == nil && status.Exit == 0 {
			return nil
		}

		combinedOutput := strings.TrimSpace(strings.Join(append(status.Stderr, status.Stdout...), "\n"))
		if combinedOutput == "" {
			if status.Error != nil {
				return fmt.Errorf("命令执行失败：%s %s：%w", name, strings.Join(args, " "), status.Error)
			}
			return fmt.Errorf("命令执行失败：%s %s，退出码 %d", name, strings.Join(args, " "), status.Exit)
		}

		if status.Error != nil {
			return fmt.Errorf("命令执行失败：%s %s：%w：%s", name, strings.Join(args, " "), status.Error, combinedOutput)
		}

		return fmt.Errorf("命令执行失败：%s %s，退出码 %d：%s", name, strings.Join(args, " "), status.Exit, combinedOutput)
	case <-time.After(timeout):
		_ = cmd.Stop()
		return fmt.Errorf("命令执行超时：%s %s", name, strings.Join(args, " "))
	}
}
