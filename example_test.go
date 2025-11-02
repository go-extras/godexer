package godexer_test

import (
	"fmt"

	"github.com/go-extras/godexer"
)

// ExampleNewWithScenario demonstrates creating and running a minimal scenario
// with default evaluator functions, and reading a variable set by the script.
func ExampleNewWithScenario() {
	const script = `
commands:
  - type: message
    stepName: hello
    description: 'Hello, {{ index . "name" }}'
  - type: variable
    stepName: set
    variable: greeting
    value: 'Hi, {{ index . "name" }}'
`

	ex, err := godexer.NewWithScenario(script, godexer.WithDefaultEvaluatorFunctions())
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	vars := map[string]any{"name": "John"}
	_ = ex.Execute(vars)

	fmt.Println("Greeting:", vars["greeting"])
	// Output:
	// Greeting: Hi, John
}
