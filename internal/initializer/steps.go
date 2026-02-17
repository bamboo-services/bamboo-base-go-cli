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
		return fmt.Errorf("target directory already exists: %s", projectDir)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check target directory: %w", err)
	}

	if err := runCommand("", "git", "clone", "--depth=1", templateRepoURL, projectDir); err != nil {
		return fmt.Errorf("clone template repository failed: %w", err)
	}

	return nil
}

func removeTemplateGitMetadata(projectDir string) error {
	gitDir := filepath.Join(projectDir, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return fmt.Errorf("remove template .git directory failed: %w", err)
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
			return fmt.Errorf("read %s failed: %w", filePath, err)
		}

		updated := bytes.ReplaceAll(content, []byte(templateModule), []byte(modulePath))
		if bytes.Equal(content, updated) {
			return nil
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("stat %s failed: %w", filePath, err)
		}

		if err := os.WriteFile(filePath, updated, fileInfo.Mode().Perm()); err != nil {
			return fmt.Errorf("write %s failed: %w", filePath, err)
		}

		return nil
	})
}

func reinitializeGitRepository(projectDir string) error {
	if err := runCommand(projectDir, "git", "init", "-b", "master"); err == nil {
		return nil
	}

	if err := runCommand(projectDir, "git", "init"); err != nil {
		return fmt.Errorf("initialize git repository failed: %w", err)
	}

	if err := runCommand(projectDir, "git", "symbolic-ref", "HEAD", "refs/heads/master"); err != nil {
		if renameErr := runCommand(projectDir, "git", "branch", "-M", "master"); renameErr != nil {
			return fmt.Errorf("set git default branch to master failed: %w", err)
		}
	}

	return nil
}

func tidyGoModule(projectDir string) error {
	if err := runCommand(projectDir, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("run go mod tidy failed: %w", err)
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
				return fmt.Errorf("%s %s failed: %w", name, strings.Join(args, " "), status.Error)
			}
			return fmt.Errorf("%s %s failed with exit code %d", name, strings.Join(args, " "), status.Exit)
		}

		if status.Error != nil {
			return fmt.Errorf("%s %s failed: %w: %s", name, strings.Join(args, " "), status.Error, combinedOutput)
		}

		return fmt.Errorf("%s %s failed with exit code %d: %s", name, strings.Join(args, " "), status.Exit, combinedOutput)
	case <-time.After(timeout):
		_ = cmd.Stop()
		return fmt.Errorf("%s %s timed out", name, strings.Join(args, " "))
	}
}
