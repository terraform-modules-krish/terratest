// Wrapper for "terraform get", which fetches any other terraform templates referenced in a module.
package terraform

import (
	"log"

	"github.com/terraform-modules-krish/terratest/shell"
)

// Call Terraform Apply on the template at "templatePath" with the given "vars"
func Get(templatePath string, logger *log.Logger) error {
	return shell.RunCommand(shell.Command { Command: "terraform", Args: []string{"get", "-update"}, WorkingDir: templatePath, Env: terraformDebugEnv }, logger)
}
