package terratest

import (
	"fmt"

	"github.com/terraform-modules-krish/terratest/log"
	"github.com/terraform-modules-krish/terratest/terraform"
)

// Apply handles all setup required for a Terraform Apply operation but does not perform a destroy operation or do any cleanup.
// The caller of this function is expected to call Destroy to clean up the Terraform template when done.
func Apply(options *TerratestOptions) (string, error) {
	logger := log.NewLogger(options.TestName)
	var output string

	terraform.Init(options.TemplatePath, logger)

	// TERRAFORM APPLY
	// Download all the Terraform modules
	logger.Println("Running terraform get...")
	err := terraform.Get(options.TemplatePath, logger)
	if err != nil {
		return output, fmt.Errorf("Failed to call terraform get successfully: %s\n", err.Error())
	}

	// Apply the Terraform template
	logger.Println("Running terraform apply...")
	if len(options.RetryableTerraformErrors) > 0 {
		output, err = terraform.ApplyAndGetOutputWithRetry(options.TemplatePath, options.Vars, options.RetryableTerraformErrors, logger)
	} else {
		output, err = terraform.ApplyAndGetOutput(options.TemplatePath, options.Vars, logger)
	}
	if err != nil {
		return output, fmt.Errorf("Failed to terraform apply: %s\n", err.Error())
	}

	return output, nil
}