// Integration tests that test cross-package functionality in AWS.
package terratest

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/terraform-modules-krish/terratest/aws"
	"github.com/terraform-modules-krish/terratest/util"
)

// This is the directory where our test fixtures are.
const fixtureDir = "./test-fixtures"

func TestUploadKeyPair(t *testing.T) {
	t.Parallel()

	// Assign randomly generated values
	region := aws.GetRandomRegion(nil)
	id := util.UniqueId()

	// Create the keypair
	keyPair, err := util.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Errorf("Failed to generate keypair: %s\n", err.Error())
	}

	// Create key in EC2
	t.Logf("Creating EC2 Keypair %s in %s...", id, region)
	defer aws.DeleteEC2KeyPair(region, id)
	aws.CreateEC2KeyPair(region, id, keyPair.PublicKey)
}

func TestTerraformApplyOnMinimalExample(t *testing.T) {
	t.Parallel()

	ro := CreateRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]string)
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	ao := NewApplyOptions()
	ao.UniqueId = rand.UniqueId
	ao.TestName = "Test - TestTerraformApplyOnMinimalExample"
	ao.TemplatePath = path.Join(fixtureDir, "minimal-example")
	ao.Vars = vars
	ao.AttemptTerraformRetry = false

	_, err = ApplyAndDestroy(ao)
	if err != nil {
		t.Fatalf("Failed to ApplyAndDestroy: %s", err.Error())
	}
}

func TestTerraformApplyOnMinimalExampleWithRetry(t *testing.T) {
	t.Parallel()

	ro := CreateRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]string)
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	ao := NewApplyOptions()
	ao.UniqueId = rand.UniqueId
	ao.TestName = "Test - TestTerraformApplyOnMinimalExampleWithRetry"
	ao.TemplatePath = path.Join(fixtureDir, "minimal-example")
	ao.Vars = vars
	ao.AttemptTerraformRetry = true

	_, err = ApplyAndDestroy(ao)
	if err != nil {
		t.Fatalf("Failed to ApplyAndDestroy: %s", err.Error())
	}
}

func TestApplyOrDestroyFailsOnTerraformError(t *testing.T) {
	t.Parallel()

	ro := CreateRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]string)
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	ao := NewApplyOptions()
	ao.UniqueId = rand.UniqueId
	ao.TestName = "Test - TestApplyOrDestroyFailsOnTerraformError"
	ao.TemplatePath = path.Join(fixtureDir, "minimal-example-with-error")
	ao.Vars = vars
	ao.AttemptTerraformRetry = true

	_, err = ApplyAndDestroy(ao)
	if err != nil {
		fmt.Printf("Received expected failure message: %s. Continuing on...", err.Error())
	} else {
		t.Fatalf("Expected a terraform apply error but ApplyAndDestroy did not return an error.")
	}
}

// Test that ApplyAndDestroy correctly retries a terraform apply when a "retryableErrorMessage" is detected. We validate
// this by scanning for a string in the output that explicitly indicates a terraform apply retry.
func TestTerraformApplyOnMinimalExampleWithRetryableErrorMessages(t *testing.T) {
	t.Parallel()

	ro := CreateRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]string)
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	ao := NewApplyOptions()
	ao.UniqueId = rand.UniqueId
	ao.TestName = "Test - TestTerraformApplyOnMinimalExampleWithRetryableErrorMessages"
	ao.TemplatePath = path.Join(fixtureDir, "minimal-example-with-error")
	ao.Vars = vars
	ao.AttemptTerraformRetry = true
	ao.RetryableTerraformErrors = make(map[string]string)
	ao.RetryableTerraformErrors["aws_instance.demo: Error launching source instance: InvalidKeyPair.NotFound"] = "This error was deliberately added to the template."

	output, err := ApplyAndDestroy(ao)
	if err != nil {
		if strings.Contains(output, "**TERRAFORM-RETRY**") {
			fmt.Println("Expected error was caught and a retry was attempted.")
		} else {
			t.Fatalf("Failed to catch expected error: %s", err.Error())
		}
	} else {
		t.Fatalf("Expected this template to have an error, but no error was thrown.")
	}

}

// Test that ApplyAndDestroy correctly avoids a retry when no "retryableErrorMessage" is detected.
func TestTerraformApplyOnMinimalExampleWithRetryableErrorMessagesDoesNotRetry(t *testing.T) {
	t.Parallel()

	ro := CreateRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]string)
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	ao := NewApplyOptions()
	ao.UniqueId = rand.UniqueId
	ao.TestName = "Test - TestTerraformApplyOnMinimalExampleWithRetryableErrorMessagesDoesNotRetry"
	ao.TemplatePath = path.Join(fixtureDir, "minimal-example-with-error")
	ao.Vars = vars
	ao.AttemptTerraformRetry = true
	ao.RetryableTerraformErrors = make(map[string]string)
	ao.RetryableTerraformErrors["I'm a message that shouldn't show up in the output"] = ""

	output, err := ApplyAndDestroy(ao)
	if err != nil {
		if strings.Contains(output, "**TERRAFORM-RETRY**") {
			t.Fatalf("Expected no terraform retry but instead a retry was attempted.")
		} else {
			fmt.Println("An error occurred and a retry was correctly avoided.")
		}
	} else {
		t.Fatalf("Expected this template to have an error, but no error was thrown.")
	}
}

func TestTerraformApplyAvoidsForbiddenRegion(t *testing.T) {
	t.Parallel()

	ro := CreateRandomResourceCollectionOptions()

	// Specify every region but us-east-1
	ro.ForbiddenRegions = []string{
		"us-west-1",
		"us-west-2",
		"eu-west-1",
		"eu-central-1",
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"sa-east-1"}

	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]string)
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	ao := NewApplyOptions()
	ao.UniqueId = rand.UniqueId
	ao.TestName = "Test - TestTerraformApplyAvoidsForbiddenRegion"
	ao.TemplatePath = path.Join(fixtureDir, "minimal-example")
	ao.Vars = vars
	ao.AttemptTerraformRetry = false

	_, err = ApplyAndDestroy(ao)
	if err != nil {
		t.Fatalf("Failed to ApplyAndDestroy: %s", err.Error())
	}
}
