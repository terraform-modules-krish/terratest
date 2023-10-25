package aws

import (
	"testing"
	"github.com/terraform-modules-krish/terratest/modules/random"
	"fmt"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"strings"
	"github.com/stretchr/testify/assert"
)

func TestSqsQueueMethods(t *testing.T) {
	t.Parallel()

	region := GetRandomRegion(t, nil, nil)
	uniqueId := random.UniqueId()
	namePrefix := fmt.Sprintf("sqs-queue-test-%s", uniqueId)

	url := CreateRandomQueue(t, region, namePrefix)
	defer deleteQueue(t, region, url)

	assert.True(t, queueExists(t, region, url))

	message := fmt.Sprintf("test-message-%s", uniqueId)
	timeoutSec := 20

	SendMessageToQueue(t, region, url, message)

	firstResponse := WaitForQueueMessage(t, region, url, timeoutSec)
	assert.NoError(t, firstResponse.Error)
	assert.Equal(t, message, firstResponse.MessageBody)

	DeleteMessageFromQueue(t, region, url, firstResponse.ReceiptHandle)

	secondResponse := WaitForQueueMessage(t, region, url, timeoutSec)
	assert.Error(t, secondResponse.Error, ReceiveMessageTimeout{QueueUrl: url, TimeoutSec: timeoutSec})
}

func queueExists(t *testing.T, region string, url string) bool {
	sqsClient, err := NewSqsClient(region)
	if err != nil {
		t.Fatal(err)
	}

	input := sqs.GetQueueAttributesInput{QueueUrl: aws.String(url)}

	if _, err := sqsClient.GetQueueAttributes(&input); err != nil {
		if strings.Contains(err.Error(), "NonExistentQueue") {
			return false
		}
		t.Fatal(err)
	}

	return true
}

func deleteQueue(t *testing.T, region string, url string) {
	DeleteQueue(t, region, url)
	assert.False(t, queueExists(t, region, url))
}