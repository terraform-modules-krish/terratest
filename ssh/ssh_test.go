package ssh

import (
	"testing"
	"github.com/terraform-modules-krish/terratest"
	"fmt"
	terralog "github.com/terraform-modules-krish/terratest/log"
	"strings"
	"log"
	"github.com/terraform-modules-krish/terratest/util"
	"time"
)

const TERRAFORM_OUTPUT_PUBLIC_IP = "example_public_ip"
const TERRAFORM_OUTPUT_PRIVATE_IP = "example_private_ip"
const EXPECTED_TEXT_FROM_SSH = "Hello World"

func TestSsh(t *testing.T) {
	t.Parallel()

	randomResourceCollection := createBaseRandomResourceCollection(t)
	terratestOptions := createTerratestOptions("TestSsh", "../test-fixtures/ssh-test", randomResourceCollection, t)
	defer terratest.Destroy(terratestOptions, randomResourceCollection)

	logger := terralog.NewLogger(terratestOptions.TestName)

	if _, err := terratest.Apply(terratestOptions); err != nil {
		t.Fatalf("Failed to apply templates: %s\n", err.Error())
	}

	if err := testSshToPublicHost(terratestOptions, randomResourceCollection, logger); err != nil {
		t.Fatalf("Failed to SSH to public host: %s\n", err.Error())
	}

	if err := testSshToPrivateHost(terratestOptions, randomResourceCollection, logger); err != nil {
		t.Fatalf("Failed to SSH to private host: %s\n", err.Error())
	}
}

// As of 6/9/16, these AWS regions do not support t2.nano instances
var REGIONS_WITHOUT_T2_NANO = []string{
	"ap-southeast-2",
}

func createBaseRandomResourceCollection(t *testing.T) *terratest.RandomResourceCollection {
	resourceCollectionOptions := terratest.NewRandomResourceCollectionOptions()
	resourceCollectionOptions.ForbiddenRegions = REGIONS_WITHOUT_T2_NANO

	randomResourceCollection, err := terratest.CreateRandomResourceCollection(resourceCollectionOptions)
	if err != nil {
		t.Fatalf("Failed to create random resource collection: %s\n", err.Error())
	}

	return randomResourceCollection
}

func createTerratestOptions(testName string, templatePath string, randomResourceCollection *terratest.RandomResourceCollection, t *testing.T) *terratest.TerratestOptions {
	terratestOptions := terratest.NewTerratestOptions()

	terratestOptions.UniqueId = randomResourceCollection.UniqueId
	terratestOptions.TemplatePath = templatePath
	terratestOptions.TestName = testName

	vpc, err := randomResourceCollection.GetDefaultVpc()
	if err != nil {
		t.Fatalf("Failed to get default VPC: %s\n", err.Error())
	}

	terratestOptions.Vars = map[string]string {
		"aws_region": randomResourceCollection.AwsRegion,
		"ami": randomResourceCollection.AmiId,
		"keypair_name": randomResourceCollection.KeyPair.Name,
		"vpc_id": vpc.Id,
		"name_prefix": fmt.Sprintf("ssh-test-%s", randomResourceCollection.UniqueId),
	}

	return terratestOptions
}

func testSshToPublicHost(terratestOptions *terratest.TerratestOptions, resourceCollection *terratest.RandomResourceCollection, logger *log.Logger) error {
	ip, err := terratest.Output(terratestOptions, TERRAFORM_OUTPUT_PUBLIC_IP)
	if err != nil {
		return err
	}

	host := Host {
		Hostname: ip,
		SshUserName: "ubuntu",
		SshKeyPair: resourceCollection.KeyPair,
	}

	_, err = util.DoWithRetry(fmt.Sprintf("SSH to %s", TERRAFORM_OUTPUT_PUBLIC_IP), 10, 30 * time.Second, logger, func() (string, error) {
		output, err := CheckSshCommand(host, fmt.Sprintf("echo '%s'", EXPECTED_TEXT_FROM_SSH), logger)

		if err != nil {
			return "", err
		}
		if ! strings.Contains(output, EXPECTED_TEXT_FROM_SSH) {
			return "", fmt.Errorf("Expected output to contain '%s' but got %s", EXPECTED_TEXT_FROM_SSH, output)
		}

		logger.Printf("Got expected output after SSHing to %s: %s", TERRAFORM_OUTPUT_PUBLIC_IP, EXPECTED_TEXT_FROM_SSH)
		return output, nil
	})

	return err
}

func testSshToPrivateHost(terratestOptions *terratest.TerratestOptions, resourceCollection *terratest.RandomResourceCollection, logger *log.Logger) error {
	publicIp, err := terratest.Output(terratestOptions, TERRAFORM_OUTPUT_PUBLIC_IP)
	if err != nil {
		return err
	}

	privateIp, err := terratest.Output(terratestOptions, TERRAFORM_OUTPUT_PRIVATE_IP)
	if err != nil {
		return err
	}

	publicHost := Host {
		Hostname: publicIp,
		SshUserName: "ubuntu",
		SshKeyPair: resourceCollection.KeyPair,
	}

	privateHost := Host {
		Hostname: privateIp,
		SshUserName: "ubuntu",
		SshKeyPair: resourceCollection.KeyPair,
	}

	_, err = util.DoWithRetry(fmt.Sprintf("SSH to %s via %s", TERRAFORM_OUTPUT_PRIVATE_IP, TERRAFORM_OUTPUT_PUBLIC_IP), 10, 30 * time.Second, logger, func() (string, error) {
		output, err := CheckPrivateSshConnection(publicHost, privateHost, fmt.Sprintf("echo '%s'", EXPECTED_TEXT_FROM_SSH), logger)

		if err != nil {
			return "", err
		}
		if ! strings.Contains(output, EXPECTED_TEXT_FROM_SSH) {
			return "", fmt.Errorf("Expected output to contain '%s' but got %s", EXPECTED_TEXT_FROM_SSH, output)
		}

		logger.Printf("Got expected output after SSHing to %s via %s: %s", TERRAFORM_OUTPUT_PRIVATE_IP, TERRAFORM_OUTPUT_PUBLIC_IP, EXPECTED_TEXT_FROM_SSH)
		return output, nil
	})

	return err
}