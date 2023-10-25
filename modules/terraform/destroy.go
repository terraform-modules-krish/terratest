package terraform

import (
	"github.com/terraform-modules-krish/terratest/modules/shell"
	"testing"
)

// Run terraform destroy with the given options and return stdout/stderr. 
func Destroy(t *testing.T, options *Options) string {
	out, err := DestroyE(t, options)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Run terraform destroy with the given options and return stdout/stderr.
func DestroyE(t *testing.T, options *Options) (string, error) {
	cmd := shell.Command{
		Command:    "terraform",
		Args:       FormatArgs(options.Vars, "destroy", "-force", "-input=false", "-lock=false"),
		WorkingDir: options.TerraformDir,
	}
	return shell.RunCommandAndGetOutputE(t, cmd)
}
