package parallel

import (
	"testing"
	"time"
	"log"
	terralog "github.com/terraform-modules-krish/terratest/log"
	"strconv"
	"fmt"
	"errors"
)

func TestParallelCounters(t *testing.T) {
	t.Parallel()

	testName := "TestParallelCounters"
	logger := terralog.NewLogger(testName)
	count := 5

	// Create a bunch of test functions, each one returning one of the integers from 1 to count. However, before
	// returning the integer, the function will sleep for count - its integer value (e.g. the function that returns
	// 0 will sleep for 5 seconds, the function that returns 2 will sleep for 3 seconds, etc). That way, if these
	// functions are really running in parallel, when we listen on our channel later on, we would expect to get the
	// numbers back in reverse order.
	counterTests := []Test {}
	for i := 0; i < count; i++ {
		sleepTime := time.Duration(count - i) * time.Second
		counterTests = append(counterTests, createCounterFunction(sleepTime, strconv.Itoa(i + 1), logger))
	}

	// Run all the counter functions in parallel
	counterResultChannel := make(chan TestResult, count)
	for i := 0; i < count; i++ {
		RunTestInParallel(counterTests[i], counterResultChannel)
	}

	// Listen on the counter channel. We should get back the values in reverse order.
	for i := 0; i < count; i++ {
		result := GetTestResult(counterResultChannel, testName, logger)
		expectedValue := strconv.Itoa(count - i)
		checkResultWithValue(result, expectedValue, logger, t)
	}
}

func TestGetTestResultReturnsExpectedResults(t *testing.T) {
	t.Parallel()

	testName := "TestGetTestResultReturnsExpectedResults"
	logger := terralog.NewLogger(testName)
	count := 5

	results := make(chan TestResult, count)
	for i := 0; i < count; i++ {
		results <- TestResult{Value: strconv.Itoa(i), Description: strconv.Itoa(i)}
	}

	for i := 0; i < count; i++ {
		result := GetTestResult(results, testName, logger)
		checkResultWithValue(result, strconv.Itoa(i), logger, t)
	}
}

func TestGetTestResultReturnsResultsAndErrors(t *testing.T) {
	t.Parallel()

	testName := "TestGetTestResultReturnsResultsAndErrors"
	logger := terralog.NewLogger(testName)
	count := 5
	errorIndex := 3

	results := make(chan TestResult, count)
	for i := 0; i < count; i++ {
		if i == errorIndex {
			errText := fmt.Sprintf("We have intentionally inserted an error in result %d", i)
			results <- TestResult{Err: errors.New(errText), Description: errText}
		} else {
			results <- TestResult{Value: strconv.Itoa(i), Description: strconv.Itoa(i)}
		}
	}

	for i := 0; i < count; i++ {
		result := GetTestResult(results, testName, logger)
		if i == errorIndex {
			checkResultWithError(result, logger, t)
		} else {
			checkResultWithValue(result, strconv.Itoa(i), logger, t)
		}
	}
}

func TestGetTestResultTimesOut(t *testing.T) {
	t.Parallel()

	testName := "TestGetTestResultTimesOut"
	logger := terralog.NewLogger(testName)

	emptyResultsChannel := make(chan TestResult, 1)
	result := GetTestResultWithTimeout(emptyResultsChannel, testName, logger, 5 * time.Second)
	checkResultWithError(result, logger, t)
}

func createCounterFunction(sleepTime time.Duration, value string, logger *log.Logger) Test {
	return func() TestResult {
		result := TestResult{Description: fmt.Sprintf("Sleeping for %s before returning value %s", sleepTime.String(), value)}
		logger.Println(result.Description)

		time.Sleep(sleepTime)
		return result.WithValue(value)
	}
}

func checkResultWithValue(result TestResult, expectedValue string, logger *log.Logger, t *testing.T) {
	if result.Err != nil {
		t.Fatalf("Did not expect result to contain an error, but found %s", result.Err.Error())
	}

	if result.Description == "" {
		t.Fatal("Expected the result to contain a description, but it was empty")
	}

	if result.Value == expectedValue {
		logger.Printf("Got expected value %s", expectedValue)
	} else {
		t.Fatalf("Expected value %s but got %s", expectedValue, result.Value)
	}
}

func checkResultWithError(result TestResult, logger *log.Logger, t *testing.T) {
	if result.Err == nil {
		t.Fatalf("Expected the result to contain an error but got nil")
	} else {
		logger.Printf("Got error message as expected: %s", result.Err.Error())
	}

	if result.Description == "" {
		t.Fatal("Expected the result to contain a description, but it was empty")
	}
}