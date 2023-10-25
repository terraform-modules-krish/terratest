package aws

import (
	"os"

	"github.com/terraform-modules-krish/terratest/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Create a new EC2 Keypair with the given name.
func CreateEC2KeyPair(awsRegion string, name string, publicKey string) error {
	log := log.NewLogger("CreateEC2KeyPair")

	svc := ec2.New(session.New(), aws.NewConfig().WithRegion(awsRegion))
	_, err := svc.Config.Credentials.Get()
	if err != nil {
		log.Fatalf("Failed to open EC2 session: %s\n", err.Error())
	}

	params := &ec2.ImportKeyPairInput{
		KeyName: aws.String(name), // Required
		PublicKeyMaterial: []byte(publicKey), // Required
		DryRun:  aws.Bool(false),
	}

	_, err = svc.ImportKeyPair(params)
	if err != nil {
		log.Printf("Failed to import EC2 keypair: %s\n", err)
		os.Exit(1)
	}

	return nil
}

// Delete an EC2 Keypair
func DeleteEC2KeyPair(awsRegion string, name string) error {
	log := log.NewLogger("DeleteEC2KeyPair")

	svc := ec2.New(session.New(), aws.NewConfig().WithRegion(awsRegion))
	_, err := svc.Config.Credentials.Get()
	if err != nil {
		log.Printf("Failed to open EC2 session: %s\n", err.Error())
		os.Exit(1)
	}

	params := &ec2.DeleteKeyPairInput{
		KeyName: aws.String(name), // Required
		DryRun:  aws.Bool(false),
	}

	_, err = svc.DeleteKeyPair(params)
	if err != nil {
		log.Printf("Failed to delete EC2 keypair: %s\n", err)
		os.Exit(1)
	}

	return nil
}
