package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-modules-krish/terratest/modules/gcp"
	"github.com/terraform-modules-krish/terratest/modules/random"
	"github.com/terraform-modules-krish/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestTerraformGcpExample(t *testing.T) {
	t.Parallel()

	// Get the Project Id to use
	projectId := gcp.GetGoogleProjectIDFromEnvVar()

	// Create all resources in the following zone
	zone := "us-east1-b"

	// Give the example bucket a unique name so we can distinguish it from any other bucket in your GCP account
	expectedBucketName := fmt.Sprintf("terratest-gcp-example-%s", strings.ToLower(random.UniqueId()))

	// Also give the example instance a unique name
	expectedInstanceName := fmt.Sprintf("terratest-gcp-example-%s", strings.ToLower(random.UniqueId()))

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../examples/terraform-gcp-example",

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"zone":          zone,
			"instance_name": expectedInstanceName,
			"bucket_name":   expectedBucketName,
		},
	}

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	//
	// TODO - we might need to sleep for a bit before running destroy in case the resources haven't
	// been fully initialized, but for now it seems to work fine.
	defer terraform.Destroy(t, terraformOptions)

	// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the value of some of the output variables
	bucketURL := terraform.Output(t, terraformOptions, "bucket_url")
	instanceID := terraform.Output(t, terraformOptions, "instance_id")

	// Verify that the new bucket url matches the expected url
	expectedURL := fmt.Sprintf("gs://%s", expectedBucketName)
	assert.Equal(t, expectedURL, bucketURL)

	// Verify that the Storage Bucket exists
	gcp.AssertStorageBucketExists(t, expectedBucketName)

	// Add a tag to the Compute Instance
	gcp.AddLabelsToInstance(t, projectId, zone, instanceID, map[string]string{"testing": "testing-tag-value2"})

	// Look up the tags for the given Instance ID
	instanceLabels := gcp.GetLabelsForComputeInstance(t, projectId, zone, instanceID)

	testingTag, containsTestingTag := instanceLabels["testing"]
	assert.True(t, containsTestingTag)
	assert.Equal(t, "testing-tag-value2", testingTag)
}
