# AGENTS Guide for `bamboo-base-go-cli`
This guide is for coding agents operating in this repository.
User instructions override this document.

## 1) Repository Context
- Project type: Go CLI installer/scaffolder.
- Module: `github.com/bamboo-services/bamboo-base-go-cli`.
- Go version: `1.25.0`.
- Entrypoint: `main.go`.
- Core packages: `internal/cli`, `internal/initializer`.
- CLI framework: `github.com/spf13/cobra`.
- TUI framework: `github.com/charmbracelet/bubbletea`.

## 2) Cursor/Copilot Rule Files
Checked paths:
- `.cursor/rules/**/*`
- `.cursorrules`
- `.github/copilot-instructions.md`
Current state: none exist in this repository.
If these files are added later, treat them as higher-priority local rules and merge with this guide.

## 3) Build / Run / Test / Lint Commands
This repo currently has no `Makefile`, no CI workflow, and no `*_test.go` files.
Use standard Go commands.

### Build
```bash
go build ./...
go build -o bamboo .
```

### Run
```bash
go run . --help
go run . init github.com/your-org/your-project
./bamboo init github.com/your-org/your-project
```

### Test (all)
```bash
go test ./...
go test -v ./...
```

### Test (single test / single package / subtest)
```bash
go test ./... -run '^TestName$' -v
go test ./internal/initializer -run '^TestName$' -v
go test ./... -run '^TestSuite$/^subcase$' -v
```

### Coverage
```bash
go test ./... -cover
go test ./... -coverprofile=coverage.out
```

### Format / Static Analysis / Module Hygiene
```bash
go fmt ./...
go vet ./...
go mod tidy
```

### Optional Lint
`golangci-lint` is not configured in this repo.
Run only if user asks or your environment requires it.
```bash
golangci-lint run
```

## 4) Code Style Guidelines (Observed)
These rules are inferred from current source files and should be preserved.

### Imports
- Group imports as: stdlib, blank line, third-party/internal.
- Use aliases only when disambiguation is needed (for example `tea`, `gocmd`).

### Formatting
- Keep code `gofmt` clean.
- Do not manually align spacing beyond `gofmt` output.
- Prefer ASCII source text unless file context requires Unicode.

### Package / File Organization
- Keep `main.go` minimal and delegate logic to `internal/*`.
- Keep package names aligned with directory names.
- Keep files focused by concern (`cli.go`, `init.go`, `model.go`, `steps.go`).

### Naming
- Exported entrypoints use clear verbs, e.g. `Run`.
- Internal helpers use lowerCamelCase, e.g. `newState`, `validateModulePath`.
- Use short receiver names, e.g. `m` for model methods.
- Unexported constants should be descriptive lowerCamelCase.

### Error Handling
- Prefer explicit error returns; avoid `panic` for expected failures.
- Wrap underlying errors with `%w`.
- Include operation context in error messages.
- Validate input early and return quickly on failure.
- Keep error strings concise and lowercase.

### Control Flow
- Prefer guard clauses and early returns.
- Keep nesting shallow.
- Keep each function focused on one responsibility.

### CLI/TUI Behavior
- CLI argument parsing/dispatch is implemented with Cobra in `internal/cli`.
- Prefer Cobra `RunE` for command handlers so errors can be returned and surfaced in `main.go`.
- Use Cobra arg validators (for example `cobra.ExactArgs(1)`) instead of manual arg length checks.
- Initialization workflow orchestration belongs in `internal/initializer`.
- Bubble Tea flow should follow `Init` / `Update` / `View` semantics.

### External Command Execution
- Use shared helper(s) for subprocess execution (`runCommand`), not ad-hoc process wiring.
- Preserve timeout behavior for long-running subprocesses.
- Include command + args in failure context.

## 5) Agent Editing Rules
- Keep changes minimal and localized.
- Do not refactor unrelated code while fixing a bug.
- Preserve public signatures unless user asks for API change.
- Avoid adding dependencies unless clearly justified.
- If dependencies change, run `go mod tidy`.

## 6) Verification Checklist
Run after edits:
```bash
go test ./...
go build ./...
go mod tidy
```
For initializer-flow changes, also run:
```bash
go run . init github.com/example/smoke-test
```
Then verify:
- Generated project directory exists.
- Generated `go.mod` module path is rewritten.
- New git repository default branch is `master`.

## 7) References
- Testing package: https://pkg.go.dev/testing
- `go test` flags: https://pkg.go.dev/cmd/go#hdr-Testing_flags
- `gofmt` blog: https://go.dev/blog/gofmt
- Effective Go formatting: https://go.dev/doc/effective_go#formatting
- Modules reference: https://go.dev/ref/mod

Keep this file updated as repository tooling evolves (tests, CI, lint config, scripts).
