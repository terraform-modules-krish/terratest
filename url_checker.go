package terratest

import (
	"errors"
	"time"
	"github.com/terraform-modules-krish/terratest/log"
	"github.com/terraform-modules-krish/terratest/http"
)

// A helper function to check if a URL returns a 200 OK with the expectedResponse as a body. This function reads the
// domain and port as Terraform outputs using the domainKeyname and portKeyname, respectively, builds a URL from them,
// and tries the URL up to 10 times, waiting for 30 seconds between retries.
func CheckTerraformOutputUrlReturnsExpectedText(terratestOptions *TerratestOptions, domainKeyname string, portKeyname string, expectedResponse string) error {
	return CheckTerraformOutputUrlReturnsExpectedTextWithinTimeLimit(terratestOptions, domainKeyname, portKeyname, expectedResponse, 10, 30 * time.Second)
}

// A helper function to check if a URL returns a 200 OK with the expectedResponse as a body. This function reads the
// domain and port as Terraform outputs using the domainKeyname and portKeyname, respectively, builds a URL from them,
// and tries the URL up to maxRetries times, waiting for sleepBetweenRetries between retries.
func CheckTerraformOutputUrlReturnsExpectedTextWithinTimeLimit(terratestOptions *TerratestOptions, domainKeyname string, portKeyname string, expectedResponse string, maxRetries int, sleepBetweenRetries time.Duration) error {
	domain, err := Output(terratestOptions, domainKeyname)
	if err != nil {
		return err
	}
	if domain == "" {
		return errors.New("Got empty value for Terraform output " + domainKeyname)
	}

	port, err := Output(terratestOptions, portKeyname)
	if err != nil {
		return err
	}
	if port == "" {
		return errors.New("Got empty value for Terraform output " + portKeyname)
	}

	url := "http://" + domain + ":" + port + "/"
	return http_helper.HttpGetWithRetry(url, expectedResponse, maxRetries, sleepBetweenRetries, log.NewLogger(terratestOptions.TestName))
}
