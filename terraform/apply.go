// Wrapper functions for "terraform apply"
package terraform

import (
	"log"
	"strings"

	"github.com/terraform-modules-krish/terratest/shell"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// NOTE: Copy & pasted from Jim's original terraform-modules/test/terraform_helper.go
// - Main modifications are standardizing and otherwise prettying up the code
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// To debug or report Terraform bugs, you need to add a TF_LOG environment variable and set it to a log level like
// TRACE (https://www.terraform.io/docs/internals/debugging.html). However, it seems that whenever you enable any
// logging with TF_LOG, Terraform spams a TON of logging at you, which makes it hard to read the output. Therefore,
// we pass an empty env var map for now, but if you need to debug, just uncomment the version with TF_LOG.
//var terraformDebugEnv = map[string]string{"TF_LOG": "INFO"}
var terraformDebugEnv = map[string]string{}

// Call Terraform Apply on the template at "templatePath" with the given "vars"
func Apply(templatePath string, vars map[string]interface{}, logger *log.Logger) error {
	return shell.RunCommand(shell.Command { Command: "terraform", Args: FormatArgs(vars, "apply", "-input=false", "-lock=false", "-auto-approve"), WorkingDir: templatePath, Env: terraformDebugEnv }, logger)
}

// Call Terraform Apply on the template at "templatePath" with the given "vars", and return the output as a string
func ApplyAndGetOutput(terraformPath string, vars map[string]interface{}, logger *log.Logger) (string, error) {
	logger.Println("Applying Terraform templates in folder", terraformPath)
	return shell.RunCommandAndGetOutput( shell.Command { Command: "terraform", Args: FormatArgs(vars, "apply", "-input=false", "-lock=false", "-auto-approve"), WorkingDir: terraformPath, Env: terraformDebugEnv }, logger)
}

// Regrettably Terraform has many bugs. Often, just re-running terraform apply will resolve the issue.
// This function declares which Terraform error messages warrant an automatic retry and does the retry.
func ApplyAndGetOutputWithRetry(terraformPath string, vars map[string]interface{}, errors map[string]string, logger *log.Logger) (string, error) {
	output, err := ApplyAndGetOutput(terraformPath, vars, logger)
	if err != nil {
		logger.Printf("Terraform apply failed with error: %s\n", err.Error())

		// Check for terraform errors that apply to all terraform templates.
		if strings.Contains(output, TF_ERROR_DIFFS_DIDNT_MATCH_DURING_APPLY) {
			logger.Printf("Terraform apply failed with the error '%s'. %s\n", TF_ERROR_DIFFS_DIDNT_MATCH_DURING_APPLY, TF_ERROR_DIFFS_DIDNT_MATCH_DURING_APPLY_MSG)
			retryOutput, err := ApplyAndGetOutput(terraformPath, vars, logger)
			return output + "**TERRAFORM-RETRY**\n" + retryOutput, err
		}

		// Check for terraform errors that are specific to this template.
		for errorText, errorTextMsg := range errors {
			if strings.Contains(output, errorText) {
				logger.Printf("Terraform apply failed with the error '%s' but this error was expected and warrants a terraform apply retry. Further details: %s\n", errorText, errorTextMsg)
				retryOutput, err := ApplyAndGetOutput(terraformPath, vars, logger)
				return output + "**TERRAFORM-RETRY**\n" + retryOutput, err

			}
		}

		logger.Printf("Terraform failed with an error we didn't expect: %s", err.Error())
		return output, err
	}

	return output, nil
}