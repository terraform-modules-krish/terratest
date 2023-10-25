package util

import (
	"testing"
	"time"
	terralog "github.com/terraform-modules-krish/terratest/log"
	"fmt"
)

func TestDoWithRetry(t *testing.T) {
	t.Parallel()

	expectedOutput := "expected"
	expectedError := fmt.Errorf("expected error")

	actionAlwaysReturnsExpected := func() (string, error) { return expectedOutput, nil }
	actionAlwaysReturnsError := func() (string, error) { return "", expectedError }

	createActionThatReturnsExpectedAfterFiveRetries := func() func() (string, error) {
		count := 0
		return func() (string, error) {
			count++
			if count > 5 {
				return expectedOutput, nil
			} else {
				return "", expectedError
			}
		}
	}

	testCases := []struct {
		description         string
		maxRetries          int
		expectedError       error
		action              func() (string, error)
	}{
		{"Return value on first try", 10, nil, actionAlwaysReturnsExpected},
		{"Return error on all retries", 10, MaxRetriesExceeded{Description: "Return error on all retries", MaxRetries: 10}, actionAlwaysReturnsError},
		{"Return value after 5 retries", 10, nil, createActionThatReturnsExpectedAfterFiveRetries()},
		{"Return value after 5 retries, but only do 4 retries", 4, MaxRetriesExceeded{Description: "Return value after 5 retries, but only do 4 retries", MaxRetries: 4}, createActionThatReturnsExpectedAfterFiveRetries()},
	}

	logger := terralog.NewLogger("TestDoWithRetry")

	for _, testCase := range testCases {
		actualOutput, err := DoWithRetry(testCase.description, testCase.maxRetries, 1 * time.Millisecond, logger, testCase.action)
		if testCase.expectedError != nil {
			if err != testCase.expectedError {
				t.Fatalf("Expected error '%v' for test case '%s' but got '%v'", testCase.description, testCase.expectedError, err)
			}
		} else {
			if err != nil {
				t.Fatalf("Did not expect an error for test case '%s' but got: %s", testCase.description, err.Error())
			}
			if actualOutput != expectedOutput {
				t.Fatalf("Expected output '%s' but got '%s'", expectedOutput, actualOutput)
			}
		}
	}
}

func TestDoWithTimeout(t *testing.T) {
	t.Parallel()

	expectedOutput := "expected"
	expectedError := fmt.Errorf("expected error")

	actionReturnsValueImmediately := func() (string, error) { return expectedOutput, nil }
	actionReturnsErrorImmediately := func() (string, error) { return "", expectedError}

	createActionThatReturnsValueAfterDelay := func(delay time.Duration) func() (string, error) {
		return func() (string, error) {
			time.Sleep(delay)
			return expectedOutput, nil
		}
	}

	createActionThatReturnsErrorAfterDelay := func(delay time.Duration) func() (string, error) {
		return func() (string, error) {
			time.Sleep(delay)
			return "", expectedError
		}
	}

	testCases := []struct {
		description         string
		timeout             time.Duration
		expectedError       error
		action              func() (string, error)
	}{
		{"Returns value immediately", 5 * time.Second, nil, actionReturnsValueImmediately},
		{"Returns error immediately", 5 * time.Second, expectedError, actionReturnsErrorImmediately},
		{"Returns value after 2 seconds", 5 * time.Second, nil, createActionThatReturnsValueAfterDelay(2 * time.Second)},
		{"Returns error after 2 seconds", 5 * time.Second, expectedError, createActionThatReturnsErrorAfterDelay(2 * time.Second)},
		{"Returns value after timeout exceeded", 5 * time.Second, TimeoutExceeded{Description: "Returns value after timeout exceeded", Timeout: 5 * time.Second}, createActionThatReturnsValueAfterDelay(10 * time.Second)},
		{"Returns error after timeout exceeded", 5 * time.Second, TimeoutExceeded{Description: "Returns error after timeout exceeded", Timeout: 5 * time.Second}, createActionThatReturnsErrorAfterDelay(10 * time.Second)},
	}

	for _, testCase := range testCases {
		actualOutput, err := DoWithTimeout(testCase.description, testCase.timeout, testCase.action)
		if testCase.expectedError != nil {
			if err != testCase.expectedError {
				t.Fatalf("Expected error '%v' for test case '%s' but got '%v'", testCase.description, testCase.expectedError, err)
			}
		} else {
			if err != nil {
				t.Fatalf("Did not expect an error for test case '%s' but got: %s", testCase.description, err.Error())
			}
			if actualOutput != expectedOutput {
				t.Fatalf("Expected output '%s' but got '%s'", expectedOutput, actualOutput)
			}
		}
	}
}
