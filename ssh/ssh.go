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

func CheckSshConnection(host string, user string, keyPair *terratest.Ec2Keypair, logger *log.Logger) error {
	defer cleanupKeyPairFile(keyPair, logger)
	writeKeyPairFile(keyPair, logger)

	sshErr := shell.RunCommand(shell.Command{Command: "ssh", Args: []string{"-i", keyPair.Name, "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", user + "@" + host, "'exit'"}}, logger)

	exitCode, err := shell.GetExitCodeForRunCommandError(sshErr)

	if err != nil {
		return err
	}

	if exitCode != 0 {
		return errors.New("SSH exited with a non-zero exit code: " + strconv.Itoa(exitCode))
	}

	return nil
}

// CheckPrivateSshConnection attempts to connect to a private server (i.e. not addressable from the Internet) via a separate public server (i.e. addressable from the Internet)
// It is useful for checking that it's possible to SSH from a Bastion Host to a private instance.
func CheckPrivateSshConnection(publicHost string, publicHostUser string, publicHostKeyPair *terratest.Ec2Keypair, privateHost string, privateHostUser string, privateHostKeyPair *terratest.Ec2Keypair, logger *log.Logger) error {
	defer cleanupKeyPairFile(publicHostKeyPair, logger)
	writeKeyPairFile(publicHostKeyPair, logger)

	// We need the SSH key to be available when we SSH from the Bastion Host to the Private Host.
	// We cannot guarantee ssh-agent will be in the test environment, so we use scp to copy the key to the bastion host file system.
	// Start by setting permissions on the key to 0600.
	chmodErr := shell.RunCommand(shell.Command{Command: "chmod", Args: []string{"0600", publicHostKeyPair.Name}}, logger)
	exitCode, err := shell.GetExitCodeForRunCommandError(chmodErr)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return errors.New("Attempt to set permissions on local key file exited with a non-zero exit code: " + strconv.Itoa(exitCode))
	}

	// Upload the key to the bastion host
	sshErr := shell.RunCommand(shell.Command{Command: "scp", Args: []string{"-p", "-i", publicHostKeyPair.Name, "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", publicHostKeyPair.Name, publicHostUser + "@" + publicHost + ":key.pem"}}, logger)
	exitCode, err = shell.GetExitCodeForRunCommandError(sshErr)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return errors.New("Attempt to SSH and write key file exited with a non-zero exit code: " + strconv.Itoa(exitCode))
	}

	// Now connect directly to the privateHost
	sshErr = shell.RunCommand(shell.Command{Command: "ssh", Args: []string{"-i", publicHostKeyPair.Name, "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", publicHostUser + "@" + publicHost, "ssh -i key.pem -o StrictHostKeyChecking=no", privateHostUser + "@" + privateHost}}, logger)
	exitCode, err = shell.GetExitCodeForRunCommandError(sshErr)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return errors.New("Attempt to SSH to private host exited with a non-zero exit code: " + strconv.Itoa(exitCode))
	}

	return nil
}

func writeKeyPairFile(keyPair *terratest.Ec2Keypair, logger *log.Logger) error {
	logger.Println("Creating test-time Key Pair file", keyPair.Name)
	return ioutil.WriteFile(keyPair.Name, []byte(keyPair.PrivateKey), 0400)
}

func cleanupKeyPairFile(keyPair *terratest.Ec2Keypair, logger *log.Logger) error {
	logger.Println("Cleaning up test-time Key Pair file", keyPair.Name)
	return os.Remove(keyPair.Name)
}