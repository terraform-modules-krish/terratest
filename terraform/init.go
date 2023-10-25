// Terraform doesn't haven an "init" subcommand, but it does have a "terraform remote" command which configures remote state.
// For our purposes, we want all our terraform runs to store state remotely, so we treat this as an "initialization" step
package terraform

import (
	"log"

	"github.com/terraform-modules-krish/terratest/shell"
)

func Init(templatePath string, logger *log.Logger) error {
	logger.Println("Running terraform init")
	args := []string{"init"}
	return shell.RunCommand(shell.Command{ Command: "terraform", Args: args, WorkingDir: templatePath, Env: terraformDebugEnv }, logger)
}
