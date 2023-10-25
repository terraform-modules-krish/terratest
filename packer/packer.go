package packer

import (
	"log"
	"github.com/terraform-modules-krish/terratest/shell"
	"regexp"
	"errors"
)

type PackerOptions struct {
	Template string		    // The path to the Packer template
	Vars     map[string]string  // The custom vars to pass when running the build command
	Only     string             // If specified, only run the build of this name
}

// Build the given Packer template and return the generated AMI ID
func BuildAmi(options PackerOptions, logger *log.Logger) (string, error) {
	logger.Printf("Running Packer to generate AMI for template %s", options.Template)

	output, err := shell.RunCommandAndGetOutput(shell.Command {Command: "packer", Args: formatPackerArgs(options)}, logger)
	if err != nil {
		return "", err
	}

	return extractAmiId(output)
}

// The Packer machine-readable log output should contain an entry of this format:
//
// <timestamp>,<builder>,artifact,<index>,id,<region>:<ami_id>
//
// For example:
//
// 1456332887,amazon-ebs,artifact,0,id,us-east-1:ami-b481b3de
func extractAmiId(packerLogOutput string) (string, error) {
	re := regexp.MustCompile(".+artifact,\\d+?,id,.+?:(.+)")
	matches := re.FindStringSubmatch(packerLogOutput)

	if len(matches) == 2 {
		return matches[1], nil
	} else {
		return "", errors.New("Could not find AMI ID pattern in Packer output")
	}
}

// Convert the inputs to a format palatable to packer. The build command should have the format:
//
// packer build [OPTIONS] template
func formatPackerArgs(options PackerOptions) []string {
	args := []string{"build", "-machine-readable"}

	for key, value := range options.Vars {
		args = append(args, "-var", key + "=" + value)
	}

	if options.Only != "" {
		args = append(args, "-only=" + options.Only)
	}

	return append(args, options.Template)
}

