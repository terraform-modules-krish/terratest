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

func HttpGet(url string, logger *log.Logger) (int, string, error) {
	logger.Println("Making an HTTP GET call to URL", url)

	resp, err := http.Get(url)
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
