package shell

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/stretchr/testify/require"

	"github.com/terraform-modules-krish/terratest/modules/logger"
	"github.com/terraform-modules-krish/terratest/modules/testing"
)

// Command is a simpler struct for defining commands than Go's built-in Cmd.
type Command struct {
	Command    string            // The command to run
	Args       []string          // The args to pass to the command
	WorkingDir string            // The working directory
	Env        map[string]string // Additional environment variables to set
}

// RunCommand runs a shell command and redirects its stdout and stderr to the stdout of the atomic script itself.
func RunCommand(t testing.TestingT, command Command) {
	err := RunCommandE(t, command)
	if err != nil {
		t.Fatal(err)
	}
}

// RunCommandE runs a shell command and redirects its stdout and stderr to the stdout of the atomic script itself.
func RunCommandE(t testing.TestingT, command Command) error {
	_, err := RunCommandAndGetOutputE(t, command)
	return err
}

// RunCommandAndGetOutput runs a shell command and returns its stdout and stderr as a string. The stdout and stderr of that command will also
// be printed to the stdout and stderr of this Go program to make debugging easier.
func RunCommandAndGetOutput(t testing.TestingT, command Command) string {
	out, err := RunCommandAndGetOutputE(t, command)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// RunCommandAndGetOutputE runs a shell command and returns its stdout and stderr as a string. The stdout and stderr of that command will also
// be printed to the stdout and stderr of this Go program to make debugging easier.
func RunCommandAndGetOutputE(t testing.TestingT, command Command) (string, error) {
	allOutput := []string{}
	err := runCommandAndStoreOutputE(t, command, &allOutput, &allOutput)

	output := strings.Join(allOutput, "\n")
	return output, err
}

// RunCommandAndGetStdOut runs a shell command and returns solely its stdout (but not stderr) as a string. The stdout
// and stderr of that command will also be printed to the stdout and stderr of this Go program to make debugging easier.
// If there are any errors, fail the test.
func RunCommandAndGetStdOut(t testing.TestingT, command Command) string {
	output, err := RunCommandAndGetStdOutE(t, command)
	require.NoError(t, err)
	return output
}

// RunCommandAndGetStdOutE runs a shell command and returns solely its stdout (but not stderr) as a string. The stdout
// and stderr of that command will also be printed to the stdout and stderr of this Go program to make debugging easier.
func RunCommandAndGetStdOutE(t testing.TestingT, command Command) (string, error) {
	stdout := []string{}
	stderr := []string{}
	err := runCommandAndStoreOutputE(t, command, &stdout, &stderr)

	output := strings.Join(stdout, "\n")
	return output, err
}

// runCommandAndStoreOutputE runs a shell command and stores each line from stdout and stderr in the given
// storedStdout and storedStderr variables, respectively. The stdout and stderr of that command will also
// be printed to the stdout and stderr of this Go program to make debugging easier.
func runCommandAndStoreOutputE(t testing.TestingT, command Command, storedStdout *[]string, storedStderr *[]string) error {
	logger.Logf(t, "Running command %s with args %s", command.Command, command.Args)

	cmd := exec.Command(command.Command, command.Args...)
	cmd.Dir = command.WorkingDir
	cmd.Stdin = os.Stdin
	cmd.Env = formatEnvVars(command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	if err := readStdoutAndStderr(t, stdout, stderr, storedStdout, storedStderr); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// This function captures stdout and stderr into the given variables while still printing it to the stdout and stderr
// of this Go program
func readStdoutAndStderr(t testing.TestingT, stdout io.ReadCloser, stderr io.ReadCloser, storedStdout *[]string, storedStderr *[]string) error {
	stdoutReader := bufio.NewReader(stdout)
	stderrReader := bufio.NewReader(stderr)

	wg := &sync.WaitGroup{}
	mutex := &sync.Mutex{}

	wg.Add(2)
	var stdoutErr, stderrErr error
	go func() {
		defer wg.Done()
		stdoutErr = readData(t, stdoutReader, mutex, storedStdout)
	}()
	go func() {
		defer wg.Done()
		stderrErr = readData(t, stderrReader, mutex, storedStderr)
	}()
	wg.Wait()

	if stdoutErr != nil {
		return stdoutErr
	}
	if stderrErr != nil {
		return stderrErr
	}

	return nil
}

func readData(t testing.TestingT, reader *bufio.Reader, mutex *sync.Mutex, allOutput *[]string) error {
	var line string
	var err error
	for {
		line, err = reader.ReadString('\n')

		// remove newline, our output is in a slice,
		// one element per line.
		line = strings.TrimSuffix(line, "\n")

		// only return early if the line does not have
		// any contents. We could have a line that does
		// not not have a newline before io.EOF, we still
		// need to add it to the output.
		if len(line) == 0 && err == io.EOF {
			break
		}

		logger.Log(t, line)
		mutex.Lock()
		*allOutput = append(*allOutput, line)
		mutex.Unlock()

		if err != nil {
			break
		}
	}
	if err != io.EOF {
		return err
	}
	return nil
}

// GetExitCodeForRunCommandError tries to read the exit code for the error object returned from running a shell command. This is a bit tricky to do
// in a way that works across platforms.
func GetExitCodeForRunCommandError(err error) (int, error) {
	// http://stackoverflow.com/a/10385867/483528
	if exitErr, ok := err.(*exec.ExitError); ok {
		// The program has exited with an exit code != 0

		// This works on both Unix and Windows. Although package
		// syscall is generally platform dependent, WaitStatus is
		// defined for both Unix and Windows and in both cases has
		// an ExitStatus() method with the same signature.
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
		return 1, errors.New("could not determine exit code")
	}

	return 0, nil
}

func formatEnvVars(command Command) []string {
	env := os.Environ()
	for key, value := range command.Env {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	return env
}
