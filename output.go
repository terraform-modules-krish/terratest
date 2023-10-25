package terratest

import (
	"github.com/terraform-modules-krish/terratest/terraform"
	"github.com/terraform-modules-krish/terratest/log"
	"fmt"
)

func Output(options *TerratestOptions, key string) (string, error) {
	logger := log.NewLogger(options.TestName)
	return terraform.Output(options.TemplatePath, key, logger)
}

func OutputRequired(options *TerratestOptions, key string) (string, error) {
	out, err := Output(options, key)

	if err != nil {
		return "", err
	}
	if out == "" {
		return "", EmptyOutput(key)
	}

	return out, nil
}

type EmptyOutput string
func (outputName EmptyOutput) Error() string {
	return fmt.Sprintf("Required output %s was empty", string(outputName))
}