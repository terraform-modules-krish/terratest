package aws

import (
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/aws"
	"testing"
	"github.com/terraform-modules-krish/terratest/modules/logger"
)

// Create an SNS Topic and return the ARN
func CreateSnsTopic(t *testing.T, region string, snsTopicName string) string {
	out, err := CreateSnsTopicE(t, region, snsTopicName)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Create an SNS Topic and return the ARN
func CreateSnsTopicE(t *testing.T, region string, snsTopicName string) (string, error) {
	logger.Logf(t, "Creating SNS topic %s in %s", snsTopicName, region)

	snsClient, err := NewSnsClientE(t, region)
	if err != nil {
		return "", err
	}

	createTopicInput := &sns.CreateTopicInput{
		Name: &snsTopicName,
	}

	output, err := snsClient.CreateTopic(createTopicInput)
	if err != nil {
		return "", err
	}

	return aws.StringValue(output.TopicArn), err
}

// Delete an SNS Topic
func DeleteSNSTopic(t *testing.T, region string, snsTopicArn string) {
	err := DeleteSNSTopicE(t, region, snsTopicArn)
	if err != nil {
		t.Fatal(err)
	}
}

// Delete an SNS Topic
func DeleteSNSTopicE(t *testing.T, region string, snsTopicArn string) error {
	logger.Logf(t, "Deleting SNS topic %s in %s", snsTopicArn, region)

	snsClient, err := NewSnsClientE(t, region)
	if err != nil {
		return err
	}

	deleteTopicInput := &sns.DeleteTopicInput{
		TopicArn: aws.String(snsTopicArn),
	}

	_, err = snsClient.DeleteTopic(deleteTopicInput)
	return err
}

// Create a new SNS client
func NewSnsClient(t *testing.T, region string) *sns.SNS {
	client, err := NewSnsClientE(t, region)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// Create a new SNS client
func NewSnsClientE(t *testing.T, region string) (*sns.SNS, error) {
	sess, err := NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	return sns.New(sess), nil
}