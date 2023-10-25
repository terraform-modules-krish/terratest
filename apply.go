package terratest

import (
	"fmt"

	"github.com/terraform-modules-krish/terratest/log"
	"github.com/terraform-modules-krish/terratest/aws"
	"github.com/terraform-modules-krish/terratest/terraform"
)

// Apply handles all setup required for a Terraform Apply operation but does not perform a destroy operation or do any cleanup.
// The caller of this function is expected to call Destroy to clean up the Terraform template when done.
func Apply(options *TerratestOptions) (string, error) {
	logger := log.NewLogger(options.TestName)
	var output string

	// SETUP
	// Configure terraform to use Remote State.
	err := aws.AssertS3BucketExists(options.TfRemoteStateS3BucketRegion, options.TfRemoteStateS3BucketName)
	if err != nil {
		return output, fmt.Errorf("Test failed because the S3 Bucket '%s' does not exist in the '%s' region.\n", options.TfRemoteStateS3BucketName, options.TfRemoteStateS3BucketRegion)
	}

	terraform.ConfigureRemoteState(options.TemplatePath, options.TfRemoteStateS3BucketName, options.GetTfStateFileName(), options.TfRemoteStateS3BucketRegion, logger)

	// TERRAFORM APPLY
	// Download all the Terraform modules
	logger.Println("Running terraform get...")
	err = terraform.Get(options.TemplatePath, logger)
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