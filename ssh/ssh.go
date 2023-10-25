package ssh

import (
	"github.com/terraform-modules-krish/terratest"
	"log"
	"github.com/terraform-modules-krish/terratest/shell"
	"errors"
	"strconv"
	"io/ioutil"
	"os"
	"fmt"
	"github.com/terraform-modules-krish/terratest/util"
)

type Host struct {
	Hostname string
	SshUserName string
	SshKeyPair *terratest.Ec2Keypair
}

func CheckSshConnection(host Host, logger *log.Logger) error {
	keyPairWithUniqueName := createKeyPairCopyWithUniqueName(*host.SshKeyPair)

	defer cleanupKeyPairFile(keyPairWithUniqueName, logger)
	writeKeyPairFile(keyPairWithUniqueName, logger)

	sshErr := shell.RunCommand(shell.Command{Command: "ssh", Args: []string{"-i", keyPairWithUniqueName.Name, "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", host.SshUserName + "@" + host.Hostname, "'exit'"}}, logger)

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
	publicKeyPairWithUniqueName := createKeyPairCopyWithUniqueName(*publicHost.SshKeyPair)
	privateKeyPairWithUniqueName := createKeyPairCopyWithUniqueName(*privateHost.SshKeyPair)

	defer cleanupKeyPairFile(publicKeyPairWithUniqueName, logger)
	writeKeyPairFile(publicKeyPairWithUniqueName, logger)

	defer cleanupKeyPairFile(privateKeyPairWithUniqueName, logger)
	writeKeyPairFile(privateKeyPairWithUniqueName, logger)

	// We need the SSH key to be available when we SSH from the Bastion Host to the Private Host.
	// We cannot guarantee ssh-agent will be in the test environment, so we use scp to copy the key to the bastion host file system.
	// Start by setting permissions on the key to 0600. These permissions (read/write for file owner only) are required by ssh to access the key.
	chmodErr := shell.RunCommand(shell.Command{Command: "chmod", Args: []string{"0600", privateKeyPairWithUniqueName.Name}}, logger)
	exitCode, err := shell.GetExitCodeForRunCommandError(chmodErr)
	if err != nil {
		return "", err
	}
	if exitCode != 0 {
		return "", errors.New("Attempt to set permissions on local key file exited with a non-zero exit code: " + strconv.Itoa(exitCode))
	}

	// Upload the key to the bastion host
	sshErr := shell.RunCommand(shell.Command{Command: "scp", Args: []string{"-p", "-i", publicKeyPairWithUniqueName.Name, "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", privateKeyPairWithUniqueName.Name, publicHost.SshUserName + "@" + publicHost.Hostname + ":key.pem"}}, logger)
	exitCode, err = shell.GetExitCodeForRunCommandError(sshErr)
	if err != nil {
		return "", err
	}
	if exitCode != 0 {
		return "", errors.New("Attempt to SSH and write key file exited with a non-zero exit code: " + strconv.Itoa(exitCode))
	}

	// Now connect directly to the privateHost
	output, sshErr := shell.RunCommandAndGetOutput(shell.Command{Command: "ssh", Args: []string{"-i", publicKeyPairWithUniqueName.Name, "-o", "UserKnownHostsFile=/dev/null", "-o", "StrictHostKeyChecking=no", publicHost.SshUserName + "@" + publicHost.Hostname, "ssh -i key.pem -o StrictHostKeyChecking=no", privateHost.SshUserName + "@" + privateHost.Hostname, command}}, logger)
	exitCode, err = shell.GetExitCodeForRunCommandError(sshErr)
	if err != nil {
		return output, err
	}
	if exitCode != 0 {
		return output, errors.New("Attempt to SSH to private host exited with a non-zero exit code: " + strconv.Itoa(exitCode))
	}

	return output, nil
}

func writeKeyPairFile(keyPair terratest.Ec2Keypair, logger *log.Logger) error {
	logger.Println("Creating test-time Key Pair file", keyPair.Name)
	return ioutil.WriteFile(keyPair.Name, []byte(keyPair.PrivateKey), 0400)
}

func cleanupKeyPairFile(keyPair terratest.Ec2Keypair, logger *log.Logger) error {
	logger.Println("Cleaning up test-time Key Pair file", keyPair.Name)
	return os.Remove(keyPair.Name)
}

// Testing SSH connectivity involves writing and deleting Key Pair files on disk. Since there might be multiple SSH
// checks happening in parallel, we use this function to give the Key Pair file a unique name, and thereby avoid the
// files overwriting each other.
func createKeyPairCopyWithUniqueName(keyPair terratest.Ec2Keypair) terratest.Ec2Keypair {
	// This automatically creates a shallow copy in Go
	keyPairWithUniqueName := keyPair
	keyPairWithUniqueName.Name = fmt.Sprintf("%s-%s", keyPairWithUniqueName.Name, util.UniqueId())
	return keyPairWithUniqueName
}