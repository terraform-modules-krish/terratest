package terraform

import (
	"testing"
	"github.com/terraform-modules-krish/terratest/modules/shell"
	"strings"
	"github.com/terraform-modules-krish/terratest/modules/retry"
	"github.com/terraform-modules-krish/terratest/modules/logger"
)

// Run terraform init and apply with the given options and return stdout/stderr from the apply command. Note that this
// method does NOT call destroy and assumes the caller is responsible for cleaning up any resources created by running
// apply.
func InitAndApply(t *testing.T, options *Options) string {
	out, err := InitAndApplyE(t, options)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Run terraform init and apply with the given options and return stdout/stderr from the apply command. Note that this
// method does NOT call destroy and assumes the caller is responsible for cleaning up any resources created by running
// apply.
func InitAndApplyE(t *testing.T, options *Options) (string, error) {
	if _, err := InitE(t, options); err != nil {
		return "", err
	}

	if _, err := GetE(t, options); err != nil {
		return "", err
	}

	return ApplyE(t, options)
}

// Run terraform apply with the given options and return stdout/stderr. Note that this method does NOT call destroy and
// assumes the caller is responsible for cleaning up any resources created by running apply.
func Apply(t *testing.T, options *Options) string {
	out, err := ApplyE(t, options)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Run terraform apply with the given options and return stdout/stderr. Note that this method does NOT call destroy and
// assumes the caller is responsible for cleaning up any resources created by running apply.
func ApplyE(t *testing.T, options *Options) (string, error) {
	return retry.DoWithRetryE(t, "Running terraform apply", options.MaxRetries, options.TimeBetweenRetries, func() (string, error) {
		cmd := shell.Command {
			Command:    "terraform",
			Args:       FormatArgs(options.Vars, "apply", "-input=false", "-lock=false", "-auto-approve"),
			WorkingDir: options.TerraformDir,
		}

		out, err := shell.RunCommandAndGetOutputE(t, cmd)
		if err == nil {
			return out, nil
		}

		for errorText, errorMessage := range options.RetryableTerraformErrors {
			if strings.Contains(err.Error(), errorText) {
				logger.Logf(t, "terraform apply failed with the error '%s' but this error was expected and warrants a retry. Further details: %s\n", errorText, errorMessage)
				return "", err
			}
		}

		return "", retry.FatalError{Underlying: err}
	})
}