package docker

import (
	"strconv"
	"testing"

	"github.com/terraform-modules-krish/terratest/modules/logger"
	"github.com/terraform-modules-krish/terratest/modules/shell"
	"github.com/stretchr/testify/require"
)

// StopOptions defines the options that can be passed to the 'docker stop' command
type StopOptions struct {
	// Seconds to wait for stop before killing the container (default 10)
	Time int
}

// Stop runs the 'docker stop' command for the given containers and return the stdout/stderr. This method fails
// the test if there are any errors
func Stop(t *testing.T, containers []string, options *StopOptions) string {
	out, err := StopE(t, containers, options)
	require.NoError(t, err)
	return out
}

// StopE runs the 'docker stop' command for the given containers and returns any errors.
func StopE(t *testing.T, containers []string, options *StopOptions) (string, error) {
	logger.Logf(t, "Running 'docker stop' on containers '%s'", containers)

	args, err := formatDockerStopArgs(containers, options)
	if err != nil {
		return "", err
	}

	cmd := shell.Command{
		Command: "docker",
		Args:    args,
	}

	return shell.RunCommandAndGetOutputE(t, cmd)

}

// formatDockerStopArgs formats the arguments for the 'docker stop' command
func formatDockerStopArgs(containers []string, options *StopOptions) ([]string, error) {
	args := []string{"stop"}

	if options.Time != 0 {
		args = append(args, "--time", strconv.Itoa(options.Time))
	}

	args = append(args, containers...)

	return args, nil
}
