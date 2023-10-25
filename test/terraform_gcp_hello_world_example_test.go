package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-modules-krish/terratest/modules/gcp"
	"github.com/terraform-modules-krish/terratest/modules/random"
	"github.com/terraform-modules-krish/terratest/modules/terraform"
)

func TestTerraformGcpHelloWorldExample(t *testing.T) {
	t.Parallel()

	// website::tag::1:: Get the Project Id to use
	projectId := gcp.GetGoogleProjectIDFromEnvVar(t)

	// website::tag::2:: Give the example instance a unique name
	instanceName := fmt.Sprintf("gcp-hello-world-example-%s", strings.ToLower(random.UniqueId()))

	terraformOptions := &terraform.Options{
		// website::tag::3:: The path to where our Terraform code is located
		TerraformDir: "../examples/terraform-gcp-hello-world-example",

		// website::tag::4:: Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"instance_name": instanceName,
		},

		// website::tag::5:: Variables to pass to our Terraform code using TF_VAR_xxx environment variables
		EnvVars: map[string]string{
			"GOOGLE_CLOUD_PROJECT": projectId,
		},
	}

	// website::tag::7:: At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// website::tag::6:: Run `terraform init` and `terraform apply`. Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)
}
