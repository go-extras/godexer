package main

import (
	"fmt"

	"github.com/go-extras/godexer"
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
  - type: repeat_for
    stepName: step_8
    description: Let's try repeat for map
    variable: map
    commands:
      - type: message
        stepName: step_8_message
        description: 'Step 8 Message: {{ index . "key" }}: {{ index . "value" }}'
  - type: foreach
    stepName: step_7
    description: Let's try repeat for slice
    variable: slice
    commands:
      - type: message
        stepName: step_7_message
        description: 'Step 7 Message: {{ index . "key" }}: {{ index . "value" }}'
  - type: foreach
    description: Let's try repeat for slice
    valueVar: pkg
    iterable:
      - jq
      - wget
    commands:
      - type: message
        stepName: step_8_message
        description: 'Step 8 Message: {{ .pkg }}'
`

func main() {
	exc, err := executor.NewWithScenario(
		commands,
		executor.WithDefaultEvaluatorFunctions(),
	)
	if err != nil {
		panic(err)
	}

	vars := map[string]any{
		"user":  "John",
		"slice": []string{"slice1", "slice2", "slice3"},
		"map":   map[string]string{"mapk1": "mapv1", "mapk2": "mapv2"},
	}
	err = exc.Execute(vars)
	if err != nil {
		panic(err)
	}

	fmt.Println("No errors occured")
	fmt.Printf("Variable %q has a value %q\n", "output", vars["output"])
}
