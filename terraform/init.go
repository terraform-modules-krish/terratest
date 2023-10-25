// Terraform doesn't haven an "init" subcommand, but it does have a "terraform remote" command which configures remote state.
// For our purposes, we want all our terraform runs to store state remotely, so we treat this as an "initialization" step
package terraform

import (
	"log"

	"github.com/terraform-modules-krish/terratest/shell"
)

func ConfigureRemoteState(templatePath string, s3BucketName string, tfStateFileName string, awsRegion string, logger *log.Logger) error {
	logger.Println("Setting up Terraform remote state storage in S3 bucket", s3BucketName, "with tfstate file name", tfStateFileName, "for folder", templatePath)
	args := []string{"remote", "config", "-backend=s3", "-backend-config=bucket=" + s3BucketName, "-backend-config=key=" + tfStateFileName, "-backend-config=encrypt=true", "-backend-config=region=" + awsRegion}
	return shell.RunCommand(shell.Command{ Command: "terraform", Args: args, WorkingDir: templatePath, Env: terraformDebugEnv }, logger)
}

