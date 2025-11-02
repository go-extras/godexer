# godexer

[Declarative command/script executor for Go]

[![CI](https://github.com/go-extras/godexer/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/go-extras/godexer/actions/workflows/ci.yml)
[![Lint](https://github.com/go-extras/godexer/actions/workflows/go-lint.yml/badge.svg?branch=master)](https://github.com/go-extras/godexer/actions/workflows/go-lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-extras/godexer.svg)](https://pkg.go.dev/github.com/go-extras/godexer)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-extras/godexer)](https://goreportcard.com/report/github.com/go-extras/godexer)
[![codecov](https://codecov.io/gh/go-extras/godexer/branch/master/graph/badge.svg)](https://codecov.io/gh/go-extras/godexer)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-00ADD8?logo=go)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE.md)
[![Status: Alpha](https://img.shields.io/badge/status-alpha-orange.svg)](#project-status)

## Overview

godexer is a small, extensible library to execute declarative “scripts” (pipelines) defined in YAML or JSON. It provides:
- A registry of command types (exec, message, sleep, variable, writefile, foreach) and optional extras (include, SSH)
- Templating for values and descriptions (Go text/template) with helper functions
- Conditional execution via expressions (govaluate) and pluggable evaluator functions
- Simple composition (include, foreach) and hooks-after for post-step logic

Use it to automate local/remote setup tasks, provisioning steps, bootstrapping, CI bits, or any repeatable sequence you want to keep in config but extend in Go.

Target audience: developers, SRE/DevOps engineers, and tool authors who prefer declarative workflows but need programmatic extension points.

## Features
- YAML/JSON scenarios with Go templates for values and descriptions
- Built-ins: exec, message, sleep, variable, writefile, foreach
- Conditions with `requires:` using evaluator functions
  - Defaults: `file_exists`, `strlen`, `shell_escape`
  - Version helpers via subpackage: `version_lt/lte/gt/gte/eq`
- Variables: set and consume, capture command output
- Hooks-after: register named callbacks invoked after steps
- Extensible: register your own command types and value functions
- Optional SSH commands (exec, scp writefile) via `github.com/go-extras/godexer/ssh`

## Requirements
- Go 1.25+ (CI runs on 1.25.x)

## Installation
Add the library to your module:

```sh
go get github.com/go-extras/godexer@latest
# Optional extras
go get github.com/go-extras/godexer/ssh@latest
go get github.com/go-extras/godexer/version@latest
```

## Quick start
A minimal scenario using built-ins and default evaluator functions:

```go
ex, _ := godexer.NewWithScenario(`
commands:
  - type: message
    stepName: hello
    description: 'Hello, {{ index . "name" }}'
  - type: sleep
    stepName: wait
    description: 'Sleeping a moment'
    seconds: 1
`, godexer.WithDefaultEvaluatorFunctions())
_ = ex.Execute(map[string]any{"name": "world"})
```

Practical example with `exec` capturing output and a condition:

```yaml
commands:
  - type: exec
    stepName: uname
    description: 'Capture uname'
    cmd: ["go", "version"]
    variable: go_version
  - type: message
    stepName: show
    description: 'Go version: {{ index . "go_version" }}'
    requires: 'strlen(go_version) > 0'
```

Run the runnable example in this repo:

```sh
go run ./example/local
```

See also the SSH example at `example/ssh` (requires an SSH server and key; see the file header for flags).

## API documentation
- Package reference: https://pkg.go.dev/github.com/go-extras/godexer
- Subpackages:
  - SSH commands: https://pkg.go.dev/github.com/go-extras/godexer/ssh
  - Version functions: https://pkg.go.dev/github.com/go-extras/godexer/version

## Concepts and built-ins
- Base fields (available on all commands): `type`, `stepName`, `description`, `requires`, `callsAfter`
- exec: run a process; supports env, retries (`attempts`, `delay`), `allowFail`, capture to `variable`
- message: prints description only
- sleep: pause for N seconds
- variable: set a variable from a literal or template
- writefile: write rendered contents to a file
- foreach: iterate over a slice/map; set `keyVar`/`valueVar` and run nested commands
- include (opt-in): register the `include` command by wiring a storage

Register `include` with a filesystem:

```go
cmds := godexer.GetRegisteredCommands()
cmds["include"] = godexer.NewIncludeCommandWithBasePath(os.DirFS("."), "./scripts")
ex, _ := godexer.NewWithScenario(scenario, godexer.WithCommandTypes(cmds))
```

Add version comparison helpers:

```go
ex, _ := godexer.NewWithScenario(scn, godexer.WithDefaultEvaluatorFunctions(), version.WithVersionFuncs())
```

SSH commands (exec, scp writefile):

```go
cmds := godexer.GetRegisteredCommands()
cmds["ssh_exec"] = sshexec.NewSSHExecCommand(client, os.Stdout, os.Stderr)
cmds["scp_writefile"] = sshexec.NewScpWriterFileCommand(client)
ex, _ := godexer.NewWithScenario(scn, godexer.WithCommandTypes(cmds))
```

## Contributing
Contributions are welcome! Please:
- Read CONTRIBUTING.md for details on workflow, code style, and testing
- Open issues with clear steps to reproduce
- For PRs: add/adjust tests, run `go test -race ./...`, keep `go mod tidy`, and address lint (`golangci-lint run`)

Links:
- Issues: https://github.com/go-extras/godexer/issues
- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)

## Project status
Alpha. Expect rapid changes and occasional breaking updates while the API settles.

## License
MIT © 2025 Denis Voytyuk — see [LICENSE.md](LICENSE.md).
