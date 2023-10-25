package helm

import (
	"testing"

	gruntwork-cli "github.com/terraform-modules-krish/go-commons/errors"
	"github.com/terraform-modules-krish/terratest/modules/shell"
)

// GetCommonArgs extracts common helm options. In this case, these are:
// - kubeconfig path
// - kubeconfig context
// - namespace
// - helm home path
func getCommonArgs(options *Options, args ...string) []string {
	if options.KubectlOptions != nil && options.KubectlOptions.ContextName != "" {
		args = append(args, "--kube-context", options.KubectlOptions.ContextName)
	}
	if options.KubectlOptions != nil && options.KubectlOptions.ConfigPath != "" {
		args = append(args, "--kubeconfig", options.KubectlOptions.ConfigPath)
	}
	if options.KubectlOptions != nil && options.KubectlOptions.Namespace != "" {
		args = append(args, "--namespace", options.KubectlOptions.Namespace)
	}
	if options.HomePath != "" {
		args = append(args, "--home", options.HomePath)
	}
	return args
}

// RunHelmCommandAndGetOutputE runs helm with the given arguments and options and returns stdout/stderr.
func RunHelmCommandAndGetOutputE(t *testing.T, options *Options, cmd string, additionalArgs ...string) (string, error) {
	args := []string{cmd}
	args = getCommonArgs(options, args...)

	args = append(args, formatSetValuesAsArgs(options.SetValues)...)

	valuesFilesArgs, err := formatValuesFilesAsArgsE(t, options.ValuesFiles)
	if err != nil {
		return "", errors.WithStackTrace(err)
	}
	args = append(args, valuesFilesArgs...)

	setFilesArgs, err := formatSetFilesAsArgsE(t, options.SetFiles)
	if err != nil {
		return "", errors.WithStackTrace(err)
	}
	args = append(args, setFilesArgs...)

	args = append(args, additionalArgs...)

	helmCmd := shell.Command{
		Command:    "helm",
		Args:       args,
		WorkingDir: ".",
		Env:        options.EnvVars,
	}
	return shell.RunCommandAndGetOutputE(t, helmCmd)
}
