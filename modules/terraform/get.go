package terraform

import (
	"github.com/terraform-modules-krish/terratest/modules/shell"
	"testing"
)

// Call terraform get and return stdout/stderr
func Get(t *testing.T, options *Options) string {
	out, err := GetE(t, options)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Call terraform get and return stdout/stderr
func GetE(t *testing.T, options *Options) (string, error) {
	cmd := shell.Command {
		Command: "terraform",
		Args: []string{"get", "-update"},
		WorkingDir: options.TerraformDir,
	}
	return shell.RunCommandAndGetOutputE(t, cmd)
}
