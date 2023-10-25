package ssh

import (
	"github.com/terraform-modules-krish/terratest"
	"log"
	"golang.org/x/crypto/ssh"
	"net"
)

type Host struct {
	Hostname string
	SshUserName string
	SshKeyPair *terratest.Ec2Keypair
}

// Check that you can connect via SSH to the given host
func CheckSshConnection(host Host, logger *log.Logger) error {
	_, err := CheckSshCommand(host, "'exit'", logger)
	return err
}

// Check that you can connect via SSH to the given host and run the given command. Returns the stdout/stderr.
func CheckSshCommand(host Host, command string, logger *log.Logger) (string, error) {
	authMethods, err := createAuthMethodsForHost(host)
	if err != nil {
		return "", err
	}

	hostOptions := SshConnectionOptions{
		Username: host.SshUserName,
		Address: host.Hostname,
		Port: 22,
		Command: command,
		AuthMethods: authMethods,
	}

	sshSession := &SshSession{
		Options: &hostOptions,
		JumpHost: &JumpHostSession{},
	}

	defer sshSession.Cleanup(logger)

	return runSshCommand(sshSession)
}

// CheckPrivateSshConnection attempts to connect to privateHost (which is not addressable from the Internet) via a separate
// publicHost (which is addressable from the Internet) and then executes "command" on privateHost and returns its output.
// It is useful for checking that it's possible to SSH from a Bastion Host to a private instance.
func CheckPrivateSshConnection(publicHost Host, privateHost Host, command string, logger *log.Logger) (string, error) {
	jumpHostAuthMethods, err := createAuthMethodsForHost(publicHost)
	if err != nil {
		return "", err
	}

	jumpHostOptions := SshConnectionOptions{
		Username: publicHost.SshUserName,
		Address: publicHost.Hostname,
		Port: 22,
		AuthMethods: jumpHostAuthMethods,
	}

	hostAuthMethods, err := createAuthMethodsForHost(privateHost)
	if err != nil {
		return "", err
	}

	hostOptions := SshConnectionOptions{
		Username: privateHost.SshUserName,
		Address: privateHost.Hostname,
		Port: 22,
		Command: command,
		AuthMethods: hostAuthMethods,
		JumpHost: &jumpHostOptions,
	}

	sshSession := &SshSession{
		Options: &hostOptions,
		JumpHost: &JumpHostSession{},
	}

	defer sshSession.Cleanup(logger)

	return runSshCommand(sshSession)
}

func runSshCommand(sshSession *SshSession) (string, error) {
	if err := setupSshClient(sshSession); err != nil {
		return "", err
	}

	if err := setupSshSession(sshSession); err != nil {
		return "", err
	}

	bytes, err := sshSession.Session.CombinedOutput(sshSession.Options.Command)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func setupSshClient(sshSession *SshSession) error {
	if sshSession.Options.JumpHost == nil {
		return fillSshClientForHost(sshSession)
	} else {
		return fillSshClientForJumpHost(sshSession)
	}
}

func fillSshClientForHost(sshSession *SshSession) error {
	client, err := createSshClient(sshSession.Options)

	if err != nil {
		return err
	}

	sshSession.Client = client
	return nil
}

func fillSshClientForJumpHost(sshSession *SshSession) error {
	jumpHostClient, err := createSshClient(sshSession.Options.JumpHost)
	if err != nil {
		return err
	}
	sshSession.JumpHost.JumpHostClient = jumpHostClient

	hostVirtualConn, err := jumpHostClient.Dial("tcp", sshSession.Options.ConnectionString())
	if err != nil {
		return err
	}
	sshSession.JumpHost.HostVirtualConnection = hostVirtualConn

	hostConn, hostIncomingChannels, hostIncomingRequests, err := ssh.NewClientConn(hostVirtualConn, sshSession.Options.ConnectionString(), createSshClientConfig(sshSession.Options))
	if err != nil {
		return err
	}
	sshSession.JumpHost.HostConnection = hostConn

	sshSession.Client = ssh.NewClient(hostConn, hostIncomingChannels, hostIncomingRequests)
	return nil
}

func setupSshSession(sshSession *SshSession) error {
	session, err := sshSession.Client.NewSession()
	if err != nil {
		return err
	}

	sshSession.Session = session
	return nil
}

func createSshClient(options *SshConnectionOptions) (*ssh.Client, error) {
	sshClientConfig := createSshClientConfig(options)
	return ssh.Dial("tcp", options.ConnectionString(), sshClientConfig)
}

func createSshClientConfig(hostOptions *SshConnectionOptions) *ssh.ClientConfig {
	clientConfig := &ssh.ClientConfig{
		User: hostOptions.Username,
		Auth: hostOptions.AuthMethods,
		// Do not do a host key check, as Terratest is only used for testing, not prod
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	clientConfig.SetDefaults()
	return clientConfig
}

func createAuthMethodsForHost(host Host) ([]ssh.AuthMethod, error) {
	signer, err := ssh.ParsePrivateKey([]byte(host.SshKeyPair.PrivateKey))
	if err != nil {
		return []ssh.AuthMethod{}, err
	}

	return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
}
