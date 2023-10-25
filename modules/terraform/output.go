package terraform

import (
	"fmt"
	"testing"
	"github.com/terraform-modules-krish/terratest/modules/shell"
	"strings"
)

// Call terraform output for the given variable and return its value
func Output(t *testing.T, options *Options, key string) string {
	out, err := OutputE(t, options, key)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Call terraform output for the given variable and return its value
func OutputE(t *testing.T, options *Options, key string) (string, error) {
	cmd := shell.Command {
		Command: "terraform",
		Args: []string{"output", "-no-color", key},
		WorkingDir: options.TerraformDir,
	}

	output, err := shell.RunCommandAndGetOutputE(t, cmd)

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

// Call terraform output for the given variable and return its value. If the value is empty, fail the test.
func OutputRequired(t *testing.T, options *Options, key string) string {
	out, err := OutputRequiredE(t, options, key)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Call terraform output for the given variable and return its value. If the value is empty, return an error.
func OutputRequiredE(t *testing.T, options *Options, key string) (string, error) {
	out, err := OutputE(t, options, key)

	if err != nil {
		return "", err
	}
	if out == "" {
		return "", EmptyOutput(key)
	}

	return out, nil
}

type EmptyOutput string
func (outputName EmptyOutput) Error() string {
	return fmt.Sprintf("Required output %s was empty", string(outputName))
}