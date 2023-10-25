package util

import (
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
	"os"

	"github.com/terraform-modules-krish/terratest/log"
)

type keyPair struct {
	PublicKey	string
	PrivateKey	string
}

// Generate an RSA Keypair and return the public and private keys
func GenerateRSAKeyPair(keySize int) (*keyPair, error) {
	log := log.NewLogger("GenerateRSAKeyPair")

	rsaKeyPair, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		log.Printf("Failed to generate key: %s\n", err)
		os.Exit(1)
	}

	// Extract the private key
	keyPemBlock := &pem.Block{
		Type:	"RSA PRIVATE KEY",
		Bytes:	x509.MarshalPKCS1PrivateKey(rsaKeyPair),
	}

	keyPem := string(pem.EncodeToMemory(keyPemBlock))

	// Extract the public key
	sshPubKey, err := ssh.NewPublicKey(rsaKeyPair.Public())
	if err != nil {
		log.Printf("Unable to generate new OpenSSH public key: %s\n", err.Error())
		os.Exit(1)
	}

	sshPubKeyBytes := ssh.MarshalAuthorizedKey(sshPubKey)
	sshPubKeyStr := string(sshPubKeyBytes)

	// Return
	return &keyPair{ PublicKey: sshPubKeyStr, PrivateKey: keyPem }, nil
}
