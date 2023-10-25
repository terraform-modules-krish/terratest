// Integration tests that test cross-package functionality in AWS.
package terratest

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/terraform-modules-krish/terratest/aws"
	"github.com/terraform-modules-krish/terratest/util"
	"github.com/stretchr/testify/assert"
)

// This is the directory where our test fixtures are.
const fixtureDir = "./test-fixtures"

func TestUploadKeyPair(t *testing.T) {
	t.Parallel()

	// Assign randomly generated values
	region := aws.GetRandomRegion(nil, nil)
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

	ro := NewRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]interface{})
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	options := NewTerratestOptions()
	options.UniqueId = rand.UniqueId
	options.TestName = "Test - TestTerraformApplyOnMinimalExample"
	options.TemplatePath = path.Join(fixtureDir, "minimal-example")
	options.Vars = vars

	_, err = ApplyAndDestroy(options)
	if err != nil {
		t.Fatalf("Failed to ApplyAndDestroy: %s", err.Error())
	}
}

func TestTerraformApplyOnVarTest(t *testing.T) {
	t.Parallel()

	ro := NewRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	options := NewTerratestOptions()
	options.UniqueId = rand.UniqueId
	options.TestName = "Test - TestTerraformApplyOnMinimalExample"
	options.TemplatePath = path.Join(fixtureDir, "var-test")
	options.Vars = map[string]interface{}{
		"string": "string",
		"boolean": true,
		"int": 5,
		"map": map[string]string{"foo": "bar"},
		"list": []int{1, 2, 3},
	}

	_, err = Apply(options)
	if err != nil {
		t.Fatalf("Failed to Apply: %s", err.Error())
	}

	assertTerraformOutputEqual(t, "string", "string", options)
	assertTerraformOutputEqual(t, "boolean", "true", options)
	assertTerraformOutputEqual(t, "int", "5", options)
	assertTerraformOutputEqual(t, "map", "foo = bar", options)
	assertTerraformOutputEqual(t, "list", "1,\n2,\n3", options)
}

func assertTerraformOutputEqual(t *testing.T, outputName string, expected string, options *TerratestOptions) {
	actual, err := Output(options, outputName)
	assert.NoError(t, err, "Error retrieving output %s", outputName)
	assert.Equal(t, expected, actual, "Invalid value for output %s", outputName)
}

func TestApplyOrDestroyFailsOnTerraformError(t *testing.T) {
	t.Parallel()

	ro := NewRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]interface{})
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	options := NewTerratestOptions()
	options.UniqueId = rand.UniqueId
	options.TestName = "Test - TestApplyOrDestroyFailsOnTerraformError"
	options.TemplatePath = path.Join(fixtureDir, "minimal-example-with-error")
	options.Vars = vars

	_, err = ApplyAndDestroy(options)
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

	ro := NewRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]interface{})
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	options := NewTerratestOptions()
	options.UniqueId = rand.UniqueId
	options.TestName = "Test - TestTerraformApplyOnMinimalExampleWithRetryableErrorMessages"
	options.TemplatePath = path.Join(fixtureDir, "minimal-example-with-error-2")
	options.Vars = vars
	options.RetryableTerraformErrors = make(map[string]string)
	options.RetryableTerraformErrors["aws_instance.demo: Error launching source instance: InvalidKeyPair.NotFound"] = "This error was deliberately added to the template."

	output, err := ApplyAndDestroy(options)
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

	ro := NewRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	vars := make(map[string]interface{})
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	options := NewTerratestOptions()
	options.UniqueId = rand.UniqueId
	options.TestName = "Test - TestTerraformApplyOnMinimalExampleWithRetryableErrorMessagesDoesNotRetry"
	options.TemplatePath = path.Join(fixtureDir, "minimal-example-with-error")
	options.Vars = vars
	options.RetryableTerraformErrors = make(map[string]string)
	options.RetryableTerraformErrors["I'm a message that shouldn't show up in the output"] = ""

	output, err := ApplyAndDestroy(options)
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

	ro := NewRandomResourceCollectionOptions()

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

	vars := make(map[string]interface{})
	vars["aws_region"] = rand.AwsRegion
	vars["ec2_key_name"] = rand.KeyPair.Name
	vars["ec2_instance_name"] = rand.UniqueId
	vars["ec2_image"] = rand.AmiId

	options := NewTerratestOptions()
	options.UniqueId = rand.UniqueId
	options.TestName = "Test - TestTerraformApplyAvoidsForbiddenRegion"
	options.TemplatePath = path.Join(fixtureDir, "minimal-example-avoids-forbidden-region")
	options.Vars = vars

	_, err = ApplyAndDestroy(options)
	if err != nil {
		t.Fatalf("Failed to ApplyAndDestroy: %s", err.Error())
	}
}
