package http_helper

import (
	"time"
	"net/http"
	"io/ioutil"
	"strings"
	"log"
	"github.com/terraform-modules-krish/terratest/util"
	"fmt"
)

// Perform an HTTP GET on the given URL and return the HTTP status code, body, and any error.
func HttpGet(url string, logger *log.Logger) (int, string, error) {
	logger.Println("Making an HTTP GET call to URL", url)

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

// Repeatedly perform an HTTP GET on the given URL until a 200 OK with the given body is returned or until max retries
// has been exceeded.
func HttpGetWithRetry(url string, expectedBody string, retries int, sleepBetweenRetries time.Duration, logger *log.Logger) error {
	_, err := util.DoWithRetry(fmt.Sprintf("HTTP GET to URL %s", url), retries, sleepBetweenRetries, logger, func() (string, error) {
		status, body, err := HttpGet(url, logger)

		if err != nil {
			return "", err
		} else if status != 200 {
			return "", fmt.Errorf("Expected a 200 response but got %d", status)
		} else if body != expectedBody {
			return "", fmt.Errorf("Got a 200 response, but did not get expected body. Expected: %s. Got: %s.", expectedBody, body)
		} else {
			logger.Printf("Got 200 a response from URL %s and expected body: %s\n", url, body)
			return body, nil
		}
	})

	return err
}

// Repeatedly perform an HTTP GET on the given URL until the given validation function returns true or max retries
// has been exceeded.
func HttpGetWithRetryWithCustomValidation(url string, retries int, sleepBetweenRetries time.Duration, logger *log.Logger, validateResponse func(int, string) bool) error {
	_, err := util.DoWithRetry(fmt.Sprintf("HTTP GET to URL %s", url), retries, sleepBetweenRetries, logger, func() (string, error) {
		status, body, err := HttpGet(url, logger)

		if err != nil {
			return "", err
		} else if !validateResponse(status, body) {
			return "", fmt.Errorf("Validation function did not pass for status code %d and body %s", status, body)
		} else {
			logger.Printf("Validation function passed for URL %s, with status code %d and body: %s\n", url, status, body)
			return body, nil
		}
	})

	return err
}
