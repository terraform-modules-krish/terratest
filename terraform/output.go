package terraform

import (
	"log"
	"github.com/terraform-modules-krish/terratest/shell"
	"strings"
)

func Output(templatePath string, key string, logger *log.Logger) (string, error) {
	logger.Println("Getting Terraform output for key", key, "in folder", templatePath)
	output, err := shell.RunCommandAndGetOutput(shell.Command { Command: "terraform", Args: []string{"output", "-no-color", key}, WorkingDir: templatePath }, logger)
	if err != nil {
		return "", err
	} else {
		return strings.TrimSpace(output), nil
	}
}
