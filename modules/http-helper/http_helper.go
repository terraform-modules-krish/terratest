package http_helper

import (
	"time"
	"net/http"
	"io/ioutil"
	"strings"
	"fmt"
	"testing"
	"github.com/terraform-modules-krish/terratest/modules/logger"
	"github.com/terraform-modules-krish/terratest/modules/retry"
)

// Perform an HTTP GET on the given URL and return the HTTP status code and body. If there's any error, fail the test.
func HttpGet(t *testing.T, url string) (int, string) {
	statusCode, body, err := HttpGetE(t, url)
	if err != nil {
		t.Fatal(err)
	}
	return statusCode, body
}

// Perform an HTTP GET on the given URL and return the HTTP status code, body, and any error.
func HttpGetE(t *testing.T, url string) (int, string, error) {
	logger.Logf(t, "Making an HTTP GET call to URL %s", url)

	client := http.Client{
		// By default, Go does not impose a timeout, so an HTTP connection attempt can hang for a LONG time.
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return -1, "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return -1, "", err
	}

	return resp.StatusCode, strings.TrimSpace(string(body)), nil
}

// Perform an HTTP GET on the given URL and verify that you get back the expected status code and body. If either
// doesn't match, fail the test.
func HttpGetWithValidation(t *testing.T, url string, expectedStatusCode int, expectedBody string) {
	err := HttpGetWithValidationE(t, url, expectedStatusCode, expectedBody)
	if err != nil {
		t.Fatal(err)
	}
}

// Perform an HTTP GET on the given URL and verify that you get back the expected status code and body. If either
// doesn't match, return an error.
func HttpGetWithValidationE(t *testing.T, url string, expectedStatusCode int, expectedBody string) error {
	return HttpGetWithCustomValidationE(t, url, func(statusCode int, body string) bool {
		return statusCode == expectedStatusCode && body == expectedBody
	})
}

// Perform an HTTP GET on the given URL and validate the returned status code and body using the given function.
func HttpGetWithCustomValidation(t *testing.T, url string, validateResponse func(int, string) bool) {
	err := HttpGetWithCustomValidationE(t, url, validateResponse)
	if err != nil {
		t.Fatal(err)
	}
}

// Perform an HTTP GET on the given URL and validate the returned status code and body using the given function.
func HttpGetWithCustomValidationE(t *testing.T, url string, validateResponse func(int, string) bool) error {
	statusCode, body, err := HttpGetE(t, url)

	if err != nil {
		return err
	}

	if !validateResponse(statusCode, body) {
		return ValidationFunctionFailed{Url: url, Status: statusCode, Body: body}
	}

	return nil
}

// Repeatedly perform an HTTP GET on the given URL until the given status code and body are returned or until max
// retries has been exceeded.
func HttpGetWithRetry(t *testing.T, url string, expectedStatus int, expectedBody string, retries int, sleepBetweenRetries time.Duration) {
	err := HttpGetWithRetryE(t, url, expectedStatus, expectedBody, retries, sleepBetweenRetries)
	if err != nil {
		t.Fatal(err)
	}
}

// Repeatedly perform an HTTP GET on the given URL until the given status code and body are returned or until max
// retries has been exceeded.
func HttpGetWithRetryE(t *testing.T, url string, expectedStatus int, expectedBody string, retries int, sleepBetweenRetries time.Duration) error {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("HTTP GET to URL %s", url), retries, sleepBetweenRetries, func() (string, error) {
		return "", HttpGetWithValidationE(t, url, expectedStatus, expectedBody)
	})

	return err
}

// Repeatedly perform an HTTP GET on the given URL until the given validation function returns true or max retries
// has been exceeded.
func HttpGetWithRetryWithCustomValidation(t *testing.T, url string, retries int, sleepBetweenRetries time.Duration, validateResponse func(int, string) bool) {
	err := HttpGetWithRetryWithCustomValidationE(t, url, retries, sleepBetweenRetries, validateResponse)
	if err != nil {
		t.Fatal(err)
	}
}

// Repeatedly perform an HTTP GET on the given URL until the given validation function returns true or max retries
// has been exceeded.
func HttpGetWithRetryWithCustomValidationE(t *testing.T, url string, retries int, sleepBetweenRetries time.Duration, validateResponse func(int, string) bool) error {
	_, err := retry.DoWithRetryE(t, fmt.Sprintf("HTTP GET to URL %s", url), retries, sleepBetweenRetries, func() (string, error) {
		return "", HttpGetWithCustomValidationE(t, url, validateResponse)
	})

	return err
}

type ValidationFunctionFailed struct {
	Url    string
	Status int
	Body   string
}

func (err ValidationFunctionFailed) Error() string {
	return fmt.Sprintf("Validation failed for URL %s. Response status: %d. Response body:\n%s", err.Url, err.Status, err.Body)
}
