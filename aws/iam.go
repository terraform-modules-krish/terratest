package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/defaults"
	gruntwork-cli "github.com/terraform-modules-krish/go-commons/errors"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/aws/session"
)

type PolicyDocument struct {
	Version   string
	Statement []StatementEntry
}

type StatementEntry struct {
	Effect   string
	Action   []string
	Resource string
}

func CreateAwsConfig(awsRegion string) (*aws.Config, error) {
	config := defaults.Get().Config.WithRegion(awsRegion)

	_, err := config.Credentials.Get()
	if err != nil {
		return nil, errors.WithStackTraceAndPrefix(err, "Error finding AWS credentials (did you set the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables?)")
	}

	return config, nil
}

func GetIamCurrentUserName(awsRegion string) (string, error) {

	iamClient, err := CreateIamClient(awsRegion)
	if err != nil {
		return "", err
	}

	resp, err := iamClient.GetUser(&iam.GetUserInput{})
	if err != nil {
		return "", err
	}

	return *resp.User.UserName, nil
}

func GetIamCurrentUserArn(awsRegion string) (string, error) {

	iamClient, err := CreateIamClient(awsRegion)
	if err != nil {
		return "", err
	}

	resp, err := iamClient.GetUser(&iam.GetUserInput{})
	if err != nil {
		return "", err
	}

	return *resp.User.Arn, nil
}


func CreateIamClient(awsRegion string) (*iam.IAM, error) {
	awsConfig, err := CreateAwsConfig(awsRegion)
	if err != nil {
		return nil, err
	}

	return iam.New(session.New(), awsConfig), nil
}