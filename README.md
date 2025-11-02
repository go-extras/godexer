Godexer Library
===============

Godexer is a Go library for executing command pipelines using YAML configuration syntax. The library provides a clean,
extensible framework for defining and executing various types of commands with template support, conditional execution,
and error handling.

This library is designed to run configure and run command pipelines. It uses simple YAML syntax, lets you register
custom command types, add post call hooks, use templates to insert your variables inside other variables or messages and
define requirements so that some steps could be skipped depending on the conditions.

## Basic usage

Here is a basic example:

```go
package main

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/go-extras/godexer"
	"os"
)

const commands = `
commands:
  - type: message
    stepName: step_1
    description: 'Hi, {{ index . "user" }}! Welcome to the executor example.'
  - type: sleep
    stepName: step_2
    description: Here we will sleep a little bit
    seconds: 2
  - type: password
    stepName: step_3
    description: "Let's generate a password with a charset [abcdefgh]"
    variable: pwd
    charset: abcdefgh
  - type: message
    stepName: step_4
    description: Your password is {{ index . "pwd" }}
  - type: variable
    stepName: step_5
    description: Creating a new variable
    variable: output
    value: 'prefix-{{ index . "pwd" }}' 
  - type: exec
    stepName: step_6
    description: Let's check the current date
    cmd: ["date"]
`

func main() {
	fs := afero.NewOsFs()
	exc, err := executor.LoadExecutor(commands, nil, os.Stdout, os.Stderr, fs)
	if err != nil {
		panic(err)
	}

	vars := map[string]interface{}{
		"user": "John",
	}
	err = exc.Execute(vars)
	if err != nil {
		panic(err)
	}

	fmt.Println("No errors occurred")
	fmt.Printf("Variable %q has a value %q\n", "output", vars["output"])
}
```

## Built-in executor types

### Base

This is a base executor, which doesn't exist individually, but serves as a base for all the other commands.
It defines common arguments for all the other commands.

#### Accepted arguments

```yaml
- type: none                          # required
  stepName: no_step                   # required
  description: Base command executor  # required
  callsAfter: "n1"                    # optional
  requires: "var1 == var2"            # optional
```

- `type` is a name of a registered command executor (built-in types are listed below).
- `stepName` is a name of step. It can be used to detect if a previous step was skipped (using as a variable).
- `description` can accept a go template, where the map of vars is passed, so you can access it this way: `{{ index . "map_key" }}`.
- `callsAfter` is a registered callback hook that will be called right after the successful command execution (more details below).
- `requires` let's you skip some steps by some condition. It is an expression in a format defined by [govaluate](https://github.com/Knetic/govaluate). It can use existing variables and evaluator functions (more details below).

### Exec

This executor type runs a command on your host system.

#### Accepted arguments

```yaml
- type: exec                      # required
  stepName: exec_cmd              # required
  description: Running a command  # required
  cmd: ["some_console_command", "and", "its", "arguments..."] #required
```

- `cmd` items can accept a go template, where the map of vars is passed, so you can access it this way: `{{ index . "map_key" }}`.

### Execution process and output

When running a command, this executor type collects stdout and stderr and outputs to the destinations your define (os.Stdout and os.Stderr in the example).

If a command fails, the executor will return an error, resulting the whole pipeline to faile.

### Message

This executor is only used to output a description.

#### Accepted arguments

```yaml
- type: exec                      # required
  stepName: exec_cmd              # required
  description: Showing a message  # required
```

### Password

This executor type generates a password. You can optionally define password length and a character set.

#### Accepted arguments

```yaml
  - type: password                     # required
    stepName: step_3                   # required
    description: Generating a password # required
    variable: pwd                      # required
    charset: abcdefgh                  # optional
    length: 10                         # optional
```

- `variable` is a name of a variable which will hold the resulting password.
- `charset` is a set of characters which will be used to generate the password (optional).
- `length` is a length of the resulting password (the default value is 8, it is also the minimal value as well).

### Sleep

This executor type pauses the execution for a definite number of seconds.

#### Accepted arguments

```yaml
  - type: sleep                        # required
    stepName: step_3                   # required
    description: Sleeping              # required
    seconds: 10                        # required
```

- `seconds` is a number of seconds for which your script will be paused.

### Variable

This executor type sets a variable.

#### Accepted arguments

```yaml
  - type: variable                     # required
    stepName: step_3                   # required
    description: Sleeping              # required
    variable: somekey                  # required
    value: somevalue                   # required
```

- `variable` is a name of a variable which will hold the `value`.
- `value` is a value that will be stored in `variable`, it can also accept a go template, where the map of vars is passed,
so you can access it this way: `{{ index . "map_key" }}`.


### WriteFile

This executor type writes content to a file.

#### Accepted arguments

```yaml
  - type: writefile                         # required
    stepName: step_five                     # required
    description: Test call five             # required
    file: /some/file                        # required
    contents: 'value: {{ index . "var2" }}' # optional (empty string by default)
```

- `file` is a filename to write to.
- `contents` is contents that will be written to the file, it can also accept a go template, where the map of vars
is passed, so you can access it this way: `{{ index . "map_key" }}`.


## How to create your own command type (executor)

```go
type SomeNewCommand struct {
	executor.BaseCommand
	ParameterOne int `json:"parameterOne"`
	ParameterTwo string `json:"parameterTwo"`
}

func NewSomeNewCommand(ectx *executor.ExecutorContext) executor.Command {
	return &ExecCommand{
		BaseCommand: executor.SomeNewCommand{
			Stdout: ectx.Stdout,
			Stderr: ectx.Stderr,
		},
	}
}

// variables holds a list of existing variables to your command
// you are allowed to modify it whatever way you want
func (c *SomeNewCommand) Execute(variables map[string]interface{}) error {
    // << ... some logic ... >>
	return nil
}

executor.RegisterCommand("some_new_command", NewSomeNewCommand) 
```

Now you can use it in your yaml configuration:

```yaml
  - type: some_new_command                  # required
    stepName: step_five                     # required
    description: Some new command           # required
    parameterOne: 1                         # you define if it's required or optional
    parameterTwo: the second value          # you define if it's required or optional
```

## How to create your own evaluator functions

In addition to the existing evaluator functions (`strlen`, `file_exists`, and `shell_escape`), you can add your own ones.

Note: `strlen` returns a number (float64) that can be used in expressions (e.g., `strlen(var1) > 5`),
`file_exists` returns a boolean, and `shell_escape` returns a string.

```go
	exc, err := executor.LoadExecutor(commands, nil, os.Stdout, os.Stderr, fs)
	if err != nil {
		panic(err)
	}

	exc.RegisterEvaluatorFunction("your_function", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return false, errors.New("invalid number of arguments")
		}
		if args[0].(string) == "bar" {
			return true, nil 
		} 
		return false, nil
	})
```

Now you can use this function:

```yaml
  - type: writefile
    stepName: step_five        
    description: Test call five
    file: /some/file 
    contents: 'value'
    requires: 'your_function(var1)' # this will be true if `var1` is "bar", otherwise the step will be skipped.
```
