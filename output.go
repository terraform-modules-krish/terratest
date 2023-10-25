package terratest

import (
	"github.com/terraform-modules-krish/terratest/terraform"
	"github.com/terraform-modules-krish/terratest/log"
)

func Output(options *TerratestOptions, key string) (string, error) {
	logger := log.NewLogger(options.TestName)
	return terraform.Output(options.TemplatePath, key, logger)
}
