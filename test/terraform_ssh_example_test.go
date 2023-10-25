package test

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"os"

	"github.com/terraform-modules-krish/terratest/modules/aws"
	"github.com/terraform-modules-krish/terratest/modules/logger"
	"github.com/terraform-modules-krish/terratest/modules/random"
	"github.com/terraform-modules-krish/terratest/modules/retry"
	"github.com/terraform-modules-krish/terratest/modules/ssh"
	"github.com/terraform-modules-krish/terratest/modules/terraform"
	"github.com/terraform-modules-krish/terratest/modules/test-structure"
	"golang.org/x/crypto/ssh/agent"
)

// An example of how to test the Terraform module in examples/terraform-ssh-example using Terratest. The test also
// shows an example of how to break a test down into "stages" so you can skip stages by setting environment variables
// (e.g., skip stage "teardown" by setting the environment variable "SKIP_teardown=true"), which speeds up iteration
// when running this test over and over again locally.
func TestTerraformSshExample(t *testing.T) {
	t.Parallel()

	exampleFolder := "../examples/terraform-ssh-example"

	// At the end of the test, run `terraform destroy` to clean up any resources that were created
	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, exampleFolder)
		terraform.Destroy(t, terraformOptions)

		keyPair := test_structure.LoadEc2KeyPair(t, exampleFolder)
		aws.DeleteEC2KeyPair(t, keyPair)
	})

	// Deploy the example
	test_structure.RunTestStage(t, "setup", func() {
		terraformOptions, keyPair := configureTerraformOptions(t, exampleFolder)

		// Save the options and key pair so later test stages can use them
		test_structure.SaveTerraformOptions(t, exampleFolder, terraformOptions)
		test_structure.SaveEc2KeyPair(t, exampleFolder, keyPair)

		// This will run `terraform init` and `terraform apply` and fail the test if there are any errors
		terraform.InitAndApply(t, terraformOptions)
	})

	// Make sure we can SSH to the public Instance directly from the public Internet and the private Instance by using
	// the public Instance as a jump host
	test_structure.RunTestStage(t, "validate", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, exampleFolder)
		keyPair := test_structure.LoadEc2KeyPair(t, exampleFolder)

		testSSHToPublicHost(t, terraformOptions, keyPair)
		testSSHToPrivateHost(t, terraformOptions, keyPair)
		testSSHAgentToPublicHost(t, terraformOptions, keyPair)
		testSSHAgentToPrivateHost(t, terraformOptions, keyPair)
		testSCPToPublicHost(t, terraformOptions, keyPair)
	})

}

func configureTerraformOptions(t *testing.T, exampleFolder string) (*terraform.Options, *aws.Ec2Keypair) {
	// A unique ID we can use to namespace resources so we don't clash with anything already in the AWS account or
	// tests running in parallel
	uniqueID := random.UniqueId()

	// Give this EC2 Instance and other resources in the Terraform code a name with a unique ID so it doesn't clash
	// with anything else in the AWS account.
	instanceName := fmt.Sprintf("terratest-ssh-example-%s", uniqueID)

	// Pick a random AWS region to test in. This helps ensure your code works in all regions.
	awsRegion := aws.GetRandomRegion(t, nil, nil)

	// Create an EC2 KeyPair that we can use for SSH access
	keyPairName := fmt.Sprintf("terratest-ssh-example-%s", uniqueID)
	keyPair := aws.CreateAndImportEC2KeyPair(t, awsRegion, keyPairName)

	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: exampleFolder,

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"aws_region":    awsRegion,
			"instance_name": instanceName,
			"key_pair_name": keyPairName,
		},
	}

	return terraformOptions, keyPair
}

func testSSHToPublicHost(t *testing.T, terraformOptions *terraform.Options, keyPair *aws.Ec2Keypair) {
	// Run `terraform output` to get the value of an output variable
	publicInstanceIP := terraform.Output(t, terraformOptions, "public_instance_ip")

	// We're going to try to SSH to the instance IP, using the Key Pair we created earlier, and the user "ubuntu",
	// as we know the Instance is running an Ubuntu AMI that has such a user
	publicHost := ssh.Host{
		Hostname:    publicInstanceIP,
		SshKeyPair:  keyPair.KeyPair,
		SshUserName: "ubuntu",
	}

	// It can take a minute or so for the Instance to boot up, so retry a few times
	maxRetries := 30
	timeBetweenRetries := 5 * time.Second
	description := fmt.Sprintf("SSH to public host %s", publicInstanceIP)

	// Run a simple echo command on the server
	expectedText := "Hello, World"
	command := fmt.Sprintf("echo -n '%s'", expectedText)

	// Verify that we can SSH to the Instance and run commands
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		actualText, err := ssh.CheckSshCommandE(t, publicHost, command)

		if err != nil {
			return "", err
		}

		if strings.TrimSpace(actualText) != expectedText {
			return "", fmt.Errorf("Expected SSH command to return '%s' but got '%s'", expectedText, actualText)
		}

		return "", nil
	})
}

func testSSHToPrivateHost(t *testing.T, terraformOptions *terraform.Options, keyPair *aws.Ec2Keypair) {
	// Run `terraform output` to get the value of an output variable
	publicInstanceIP := terraform.Output(t, terraformOptions, "public_instance_ip")
	privateInstanceIP := terraform.Output(t, terraformOptions, "private_instance_ip")

	// We're going to try to SSH to the private instance using the public instance as a jump host. For both instances,
	// we are using the Key Pair we created earlier, and the user "ubuntu", as we know the Instances are running an
	// Ubuntu AMI that has such a user
	publicHost := ssh.Host{
		Hostname:    publicInstanceIP,
		SshKeyPair:  keyPair.KeyPair,
		SshUserName: "ubuntu",
	}
	privateHost := ssh.Host{
		Hostname:    privateInstanceIP,
		SshKeyPair:  keyPair.KeyPair,
		SshUserName: "ubuntu",
	}

	// It can take a minute or so for the Instance to boot up, so retry a few times
	maxRetries := 30
	timeBetweenRetries := 5 * time.Second
	description := fmt.Sprintf("SSH to private host %s via public host %s", publicInstanceIP, privateInstanceIP)

	// Run a simple echo command on the server
	expectedText := "Hello, World"
	command := fmt.Sprintf("echo -n '%s'", expectedText)

	// Verify that we can SSH to the Instance and run commands
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		actualText, err := ssh.CheckPrivateSshConnectionE(t, publicHost, privateHost, command)

		if err != nil {
			return "", err
		}

		if strings.TrimSpace(actualText) != expectedText {
			return "", fmt.Errorf("Expected SSH command to return '%s' but got '%s'", expectedText, actualText)
		}

		return "", nil
	})
}

func testSCPToPublicHost(t *testing.T, terraformOptions *terraform.Options, keyPair *aws.Ec2Keypair) {
	// Run `terraform output` to get the value of an output variable
	publicInstanceIP := terraform.Output(t, terraformOptions, "public_instance_ip")

	// We're going to try to SSH to the instance IP, using the Key Pair we created earlier, and the user "ubuntu",
	// as we know the Instance is running an Ubuntu AMI that has such a user
	publicHost := ssh.Host{
		Hostname:    publicInstanceIP,
		SshKeyPair:  keyPair.KeyPair,
		SshUserName: "ubuntu",
	}

	// It can take a minute or so for the Instance to boot up, so retry a few times
	maxRetries := 10
	timeBetweenRetries := 1 * time.Second
	description := fmt.Sprintf("SCP file to public host %s", publicInstanceIP)

	// Run a simple echo command on the server
	expectedText := "Hello, World"

	// Verify that we can SSH to the Instance and run commands
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		err := ssh.ScpFileToE(t, publicHost, os.FileMode(0644), "/tmp/test.txt", expectedText)
		if err != nil {
			return "", err
		}

		actualText, err := ssh.CheckSshCommandE(t, publicHost, "cat /tmp/test.txt")

		if err != nil {
			return "", err
		}

		if strings.TrimSpace(actualText) != expectedText {
			return "", fmt.Errorf("Expected SSH command to return '%s' but got '%s'", expectedText, actualText)
		}

		return "", nil
	})
}

func testSSHAgentToPublicHost(t *testing.T, terraformOptions *terraform.Options, keyPair *aws.Ec2Keypair) {
	// Run `terraform output` to get the value of an output variable
	publicInstanceIP := terraform.Output(t, terraformOptions, "public_instance_ip")

	// We're going to try to SSH to the instance IP, using the Key Pair we created earlier. Instead of
	// directly using the SSH key in the SSH connection, we're going to rely on an existing SSH agent that we
	// programatically emulate within this test. We're going to use the user "ubuntu" as we know the Instance
	// is running an Ubuntu AMI that has such a user
	publicHost := ssh.Host{
		Hostname:    publicInstanceIP,
		SshUserName: "ubuntu",
		SshAgent:    true,
	}

	// It can take a minute or so for the Instance to boot up, so retry a few times
	maxRetries := 30
	timeBetweenRetries := 5 * time.Second
	description := fmt.Sprintf("SSH with Agent to public host %s", publicInstanceIP)

	// Run a simple echo command on the server
	expectedText := "Hello, World"
	command := fmt.Sprintf("echo -n '%s'", expectedText)

	// Instantiate a temporary SSH agent
	socketDir, err := ioutil.TempDir("", "ssh-agent-")
	if err != nil {
		t.Fatal(err)
	}
	socketFile := filepath.Join(socketDir, "ssh_auth.sock")
	os.Setenv("SSH_AUTH_SOCK", socketFile)
	sshAgent, err := NewSSHAgent(socketDir, socketFile)
	if err != nil {
		t.Fatal(err)
	}
	defer sshAgent.Stop()

	// Create SSH key for the agent using the existing AWS SSH key pair
	block, _ := pem.Decode([]byte(keyPair.KeyPair.PrivateKey))
	pkey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	key := agent.AddedKey{PrivateKey: pkey}

	// Add SSH key to the agent
	// Retry until agent is ready or give up with a fatal error
	for i := 0; i < 15; i++ {
		var keys []*agent.Key
		keys, err = sshAgent.agent.List()
		if err != nil {
			logger.Logf(t, "Error listing SSH keys %v", err)
		}
		if len(keys) > 0 {
			logger.Logf(t, "Agent SSH keys: %v", keys)
			break
		} else {
			err = sshAgent.agent.Add(key)
			if err != nil {
				logger.Logf(t, "Error adding SSH key %v", err)
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	if err != nil {
		t.Fatal("Could not add any SSH key to the agent after several retries")
	}

	// Verify that we can SSH to the Instance and run commands
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {

		actualText, err := ssh.CheckSshCommandE(t, publicHost, command)

		if err != nil {
			return "", err
		}

		if strings.TrimSpace(actualText) != expectedText {
			return "", fmt.Errorf("Expected SSH command to return '%s' but got '%s'", expectedText, actualText)
		}

		return "", nil
	})
}

func testSSHAgentToPrivateHost(t *testing.T, terraformOptions *terraform.Options, keyPair *aws.Ec2Keypair) {
	// Run `terraform output` to get the value of an output variable
	publicInstanceIP := terraform.Output(t, terraformOptions, "public_instance_ip")
	privateInstanceIP := terraform.Output(t, terraformOptions, "private_instance_ip")

	// We're going to try to SSH to the private instance using the public instance as a jump host. Instead of
	// directly using the SSH key in the SSH connection, we're going to rely on an existing SSH agent that we
	// programatically emulate within this test. For both instances, we are using the Key Pair we created earlier,
	// and the user "ubuntu", as we know the Instances are running an Ubuntu AMI that has such a user
	publicHost := ssh.Host{
		Hostname:    publicInstanceIP,
		SshUserName: "ubuntu",
		SshAgent:    true,
	}
	privateHost := ssh.Host{
		Hostname:    privateInstanceIP,
		SshUserName: "ubuntu",
		SshAgent:    true,
	}

	// It can take a minute or so for the Instance to boot up, so retry a few times
	maxRetries := 30
	timeBetweenRetries := 5 * time.Second
	description := fmt.Sprintf("SSH with Agent to private host %s via public host %s", publicInstanceIP, privateInstanceIP)

	// Run a simple echo command on the server
	expectedText := "Hello, World"
	command := fmt.Sprintf("echo -n '%s'", expectedText)

	// Instantiate a temporary SSH agent
	socketDir, err := ioutil.TempDir("", "ssh-agent-")
	if err != nil {
		t.Fatal(err)
	}
	socketFile := filepath.Join(socketDir, "ssh_auth.sock")
	os.Setenv("SSH_AUTH_SOCK", socketFile)
	sshAgent, err := NewSSHAgent(socketDir, socketFile)
	if err != nil {
		t.Fatal(err)
	}
	defer sshAgent.Stop()

	// Create SSH key for the agent using the existing AWS SSH key pair
	block, _ := pem.Decode([]byte(keyPair.KeyPair.PrivateKey))
	pkey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	key := agent.AddedKey{PrivateKey: pkey}

	// Add SSH key to the agent
	// Retry until agent is ready or give up with a fatal error
	for i := 0; i < 15; i++ {
		var keys []*agent.Key
		keys, err = sshAgent.agent.List()
		if err != nil {
			logger.Logf(t, "Error listing SSH keys %v", err)
		}
		if len(keys) > 0 {
			logger.Logf(t, "Agent SSH keys: %v", keys)
			break
		} else {
			err = sshAgent.agent.Add(key)
			if err != nil {
				logger.Logf(t, "Error adding SSH key %v", err)
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	if err != nil {
		t.Fatal("Could not add any SSH key to the agent after several retries")
	}

	// Verify that we can SSH to the Instance and run commands
	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {

		actualText, err := ssh.CheckPrivateSshConnectionE(t, publicHost, privateHost, command)

		if err != nil {
			return "", err
		}

		if strings.TrimSpace(actualText) != expectedText {
			return "", fmt.Errorf("Expected SSH command to return '%s' but got '%s'", expectedText, actualText)
		}

		return "", nil
	})
}

type SSHAgent struct {
	stop       chan bool
	stopped    chan bool
	socketDir  string
	socketFile string
	agent      agent.Agent
	ln         net.Listener
}

// Create SSH agent, start it in background and returns control back to the main thread
func NewSSHAgent(socketDir string, socketFile string) (*SSHAgent, error) {
	var err error
	s := &SSHAgent{make(chan bool), make(chan bool), socketDir, socketFile, agent.NewKeyring(), nil}
	s.ln, err = net.Listen("unix", s.socketFile)
	if err != nil {
		return nil, err
	}
	go s.run()
	return s, nil
}

// SSH Agent listner and handler
func (s *SSHAgent) run() {
	defer close(s.stopped)
	for {
		select {
		case <-s.stop:
			return
		default:
			c, err := s.ln.Accept()
			if err != nil {
				select {
				// When s.Stop() closes the listner, s.ln.Accept() returns an error that can be ignored
				// since the agent is in stopping process
				case <-s.stop:
					return
				// When s.ln.Accept() returns a legit error, we print it and continue accepting further requests
				default:
					fmt.Errorf("Could not accept connection to agent %v", err)
					continue
				}
			} else {
				defer c.Close()
				go func(c io.ReadWriter) {
					err := agent.ServeAgent(s.agent, c)
					if err != nil {
						fmt.Errorf("Could not serve ssh agent %v", err)
					}
				}(c)
			}
		}
	}
}

// Stop and clean up SSH agent
func (s *SSHAgent) Stop() {
	close(s.stop)
	s.ln.Close()
	<-s.stopped
	os.RemoveAll(s.socketDir)
}
