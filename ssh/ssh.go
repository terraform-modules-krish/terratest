package ssh

import (
	"github.com/terraform-modules-krish/terratest"
	"log"
	"github.com/terraform-modules-krish/terratest/shell"
	"errors"
	"strconv"
	"io/ioutil"
	"os"
)

type Host struct {
	Hostname string
	SshUserName string
	SshKeyPair *terratest.Ec2Keypair
}

func CheckSshConnection(host Host, logger *log.Logger) error {
	defer cleanupKeyPairFile(host.SshKeyPair, logger)
	writeKeyPairFile(host.SshKeyPair, logger)

	sshErr := shell.RunCommand(shell.Command{Command: "ssh", Args: []string{"-i", host.SshKeyPair.Name, "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", host.SshUserName + "@" + host.Hostname, "'exit'"}}, logger)

	exitCode, err := shell.GetExitCodeForRunCommandError(sshErr)

	if err != nil {
		return err
	}

	if exitCode != 0 {
		return errors.New("SSH exited with a non-zero exit code: " + strconv.Itoa(exitCode))
	}

	return nil
}

// CheckPrivateSshConnection attempts to connect to privateHost (which is not addressable from the Internet) via a separate
// publicHost (which is addressable from the Internet) and then executes "command" on privateHost and returns its output.
// It is useful for checking that it's possible to SSH from a Bastion Host to a private instance.
func CheckPrivateSshConnection(publicHost Host, privateHost Host, command string, logger *log.Logger) (string, error) {
	defer cleanupKeyPairFile(publicHost.SshKeyPair, logger)
	writeKeyPairFile(publicHost.SshKeyPair, logger)

	defer cleanupKeyPairFile(privateHost.SshKeyPair, logger)
	writeKeyPairFile(privateHost.SshKeyPair, logger)

	// We need the SSH key to be available when we SSH from the Bastion Host to the Private Host.
	// We cannot guarantee ssh-agent will be in the test environment, so we use scp to copy the key to the bastion host file system.
	// Start by setting permissions on the key to 0600. These permissions (read/write for file owner only) are required by ssh to access the key.
	chmodErr := shell.RunCommand(shell.Command{Command: "chmod", Args: []string{"0600", privateHost.SshKeyPair.Name}}, logger)
	exitCode, err := shell.GetExitCodeForRunCommandError(chmodErr)
	if err != nil {
		return "", err
	}
	if exitCode != 0 {
		return "", errors.New("Attempt to set permissions on local key file exited with a non-zero exit code: " + strconv.Itoa(exitCode))
	}

	// Upload the key to the bastion host
	sshErr := shell.RunCommand(shell.Command{Command: "scp", Args: []string{"-p", "-i", publicHost.SshKeyPair.Name, "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", privateHost.SshKeyPair.Name, publicHost.SshUserName + "@" + publicHost.Hostname + ":key.pem"}}, logger)
	exitCode, err = shell.GetExitCodeForRunCommandError(sshErr)
	if err != nil {
		return "", err
	}
	if exitCode != 0 {
		return "", errors.New("Attempt to SSH and write key file exited with a non-zero exit code: " + strconv.Itoa(exitCode))
	}

	// Now connect directly to the privateHost
	output, sshErr := shell.RunCommandAndGetOutput(shell.Command{Command: "ssh", Args: []string{"-i", publicHost.SshKeyPair.Name, "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", publicHost.SshUserName + "@" + publicHost.Hostname, "ssh -i key.pem -o StrictHostKeyChecking=no", privateHost.SshUserName + "@" + privateHost.Hostname, command}}, logger)
	exitCode, err = shell.GetExitCodeForRunCommandError(sshErr)
	if err != nil {
		return output, err
	}
	if exitCode != 0 {
		return output, errors.New("Attempt to SSH to private host exited with a non-zero exit code: " + strconv.Itoa(exitCode))
	}

	return output, nil
}

func writeKeyPairFile(keyPair *terratest.Ec2Keypair, logger *log.Logger) error {
	logger.Println("Creating test-time Key Pair file", keyPair.Name)
	return ioutil.WriteFile(keyPair.Name, []byte(keyPair.PrivateKey), 0400)
}

func cleanupKeyPairFile(keyPair *terratest.Ec2Keypair, logger *log.Logger) error {
	logger.Println("Cleaning up test-time Key Pair file", keyPair.Name)
	return os.Remove(keyPair.Name)
}