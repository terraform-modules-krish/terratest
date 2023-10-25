package aws

import (
	"github.com/google/uuid"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"github.com/terraform-modules-krish/terratest/modules/logger"
)

// Create a new SQS queue with a random name that starts with the given prefix and return the queue URL
func CreateRandomQueue(t *testing.T, awsRegion string, prefix string) string {
	url, err := CreateRandomQueueE(t, awsRegion, prefix)
	if err != nil {
		t.Fatal(err)
	}
	return url
}

// Create a new SQS queue with a random name that starts with the given prefix and return the queue URL
func CreateRandomQueueE(t *testing.T, awsRegion string, prefix string) (string, error) {
	logger.Logf(t, "Creating randomly named SQS queue with prefix %s", prefix)

	sqsClient, err := NewSqsClient(awsRegion)
	if err != nil {
		return "", err
	}

	channel, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	channelName := fmt.Sprintf("%s-%s", prefix, channel.String())

	queue, err := sqsClient.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(channelName),
	})

	if err != nil {
		return "", err
	}

	return aws.StringValue(queue.QueueUrl), nil
}

// Delete the SQS queue with the given URL
func DeleteQueue(t *testing.T, awsRegion string, queueUrl string) {
	err := DeleteQueueE(t, awsRegion, queueUrl)
	if err != nil {
		t.Fatal(err)
	}
}

// Delete the SQS queue with the given URL
func DeleteQueueE(t *testing.T, awsRegion string, queueUrl string) error {
	logger.Logf(t, "Deleting SQS Queue %s", queueUrl)

	sqsClient, err := NewSqsClient(awsRegion)
	if err != nil {
		return err
	}

	_, err = sqsClient.DeleteQueue(&sqs.DeleteQueueInput{
		QueueUrl: aws.String(queueUrl),
	})

	return err
}

// Delete the message with the given receipt from the SQS queue with the given URL
func DeleteMessageFromQueue(t *testing.T, awsRegion string, queueUrl string, receipt string) {
	err := DeleteMessageFromQueueE(t, awsRegion, queueUrl, receipt)
	if err != nil {
		t.Fatal(err)
	}
}

// Delete the message with the given receipt from the SQS queue with the given URL
func DeleteMessageFromQueueE(t *testing.T, awsRegion string, queueUrl string, receipt string) error {
	logger.Logf(t, "Deleting message from queue %s (%s)", queueUrl, receipt)

	sqsClient, err := NewSqsClient(awsRegion)
	if err != nil {
		return err
	}

	_, err = sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		ReceiptHandle: &receipt,
		QueueUrl:      &queueUrl,
	})

	return err
}

// Send the given message to the SQS queue with the given URL
func SendMessageToQueue(t *testing.T, awsRegion string, queueUrl string, message string) {
	err := SendMessageToQueueE(t, awsRegion, queueUrl, message)
	if err != nil {
		t.Fatal(err)
	}
}

// Send the given message to the SQS queue with the given URL
func SendMessageToQueueE(t *testing.T, awsRegion string, queueUrl string, message string) error {
	logger.Logf(t, "Sending message %s to queue %s", message, queueUrl)

	sqsClient, err := NewSqsClient(awsRegion)
	if err != nil {
		return err
	}

	res, err := sqsClient.SendMessage(&sqs.SendMessageInput{
		MessageBody: &message,
		QueueUrl:    &queueUrl,
	})

	if err != nil {
		if strings.Contains(err.Error(), "AWS.SimpleQueueService.NonExistentQueue") {
			logger.Logf(t, fmt.Sprintf("WARN: Client has stopped listening on queue %s", queueUrl))
			return nil
		}
		return err
	}

	logger.Logf(t, "Message id %s sent to queue %s", aws.StringValue(res.MessageId), queueUrl)

	return nil
}

type QueueMessageResponse struct {
	ReceiptHandle string
	MessageBody   string
	Error         error
}

// Waits to receive a message from on the queueUrl. Since the API only allows us to wait a max 20 seconds for a new
// message to arrive, we must loop TIMEOUT/20 number of times to be able to wait for a total of TIMEOUT seconds
func WaitForQueueMessage(t *testing.T, awsRegion string, queueUrl string, timeout int) QueueMessageResponse {
	sqsClient, err := NewSqsClient(awsRegion)
	if err != nil {
		return QueueMessageResponse{Error: err}
	}

	cycles := timeout;
	cycleLength := 1;
	if timeout >= 20 {
		cycleLength = 20
		cycles = timeout / cycleLength
	}

	for i := 0; i < cycles; i++ {
		logger.Logf(t, "Waiting for message on %s (%ss)", queueUrl, strconv.Itoa(i*cycleLength))
		result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:              aws.String(queueUrl),
			AttributeNames:        aws.StringSlice([]string{"SentTimestamp"}),
			MaxNumberOfMessages:   aws.Int64(1),
			MessageAttributeNames: aws.StringSlice([]string{"All"}),
			WaitTimeSeconds:       aws.Int64(int64(cycleLength)),
		})

		if err != nil {
			return QueueMessageResponse{Error: err}
		}

		if len(result.Messages) > 0 {
			logger.Logf(t, "Message %s received on %s", *result.Messages[0].MessageId, queueUrl)
			return QueueMessageResponse{ReceiptHandle: *result.Messages[0].ReceiptHandle, MessageBody: *result.Messages[0].Body}
		}
	}

	return QueueMessageResponse{Error: ReceiveMessageTimeout{QueueUrl: queueUrl, TimeoutSec: timeout}}
}

// Create a new SQS client
func NewSqsClient(region string) (*sqs.SQS, error) {
	sess, err := NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	return sqs.New(sess), nil
}

type ReceiveMessageTimeout struct {
	QueueUrl   string
	TimeoutSec int
}

func (err ReceiveMessageTimeout) Error() string {
	return fmt.Sprintf("Failed to receive messages on %s within %s seconds", err.QueueUrl, strconv.Itoa(err.TimeoutSec))
}
