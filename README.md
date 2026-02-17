# bamboo-base-go-cli

`bamboo-base-go-cli` 是用于初始化 `bamboo-base-go-template` 的命令行安装器。

它会在当前目录下创建新项目，并自动完成以下步骤：

- 克隆模板仓库
- 删除模板 `.git` 历史
- 替换项目模块路径（`go.mod` 与代码引用）
- 重新初始化 Git 仓库并将默认分支设置为 `master`
- 执行 `go mod tidy`

## 功能特性

- 基于 `cobra` 的 CLI 命令结构
- 基于 `bubbletea` 的初始化进度 TUI
- 参数校验与错误提示
- 初始化流程自动化（开箱即用）

## 快速开始

### 1. 构建

```bash
go build -o bamboo .
```

### 2. 初始化项目

```bash
./bamboo init github.com/XiaoLFeng/hello
```

或直接运行：

```bash
go run . init github.com/XiaoLFeng/hello
```

执行完成后，会在当前目录生成 `hello/` 项目。

## 命令说明

### 查看帮助

```bash
go run . --help
go run . init --help
```

### 初始化命令

```bash
bamboo init <package-name>
```

示例：

```bash
bamboo init github.com/XiaoLFeng/hello
```

## 开发命令

```bash
go test ./...
go build ./...
go fmt ./...
go vet ./...
go mod tidy
```

## 依赖说明

- `github.com/spf13/cobra`
- `github.com/charmbracelet/bubbletea`
- `github.com/go-cmd/cmd`

## License

MIT
