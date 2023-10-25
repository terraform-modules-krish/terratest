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

func writeKeyPairFile(keyPair *terratest.Ec2Keypair, logger *log.Logger) error {
	logger.Println("Creating test-time Key Pair file", keyPair.Name)
	return ioutil.WriteFile(keyPair.Name, []byte(keyPair.PrivateKey), 0400)
}

func cleanupKeyPairFile(keyPair *terratest.Ec2Keypair, logger *log.Logger) error {
	logger.Println("Cleaning up test-time Key Pair file", keyPair.Name)
	return os.Remove(keyPair.Name)
}