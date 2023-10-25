package terraform

import (
	"github.com/terraform-modules-krish/terratest/modules/shell"
	"testing"
)

// Call terraform init and return stdout/stderr
func Init(t *testing.T, options *Options) string {
	out, err := InitE(t, options)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Call terraform init and return stdout/stderr
func InitE(t *testing.T, options *Options) (string, error) {
	cmd := shell.Command{
		Command: "terraform",
		Args: []string{"init"},
		WorkingDir: options.TerraformDir,
	}
	return shell.RunCommandAndGetOutputE(t, cmd)
}

