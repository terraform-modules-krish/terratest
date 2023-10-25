package aws

import (
	"github.com/aws/aws-sdk-go/aws/session"
	gruntwork-cli "github.com/terraform-modules-krish/go-commons/logging"
	"github.com/google/uuid"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"strconv"
	"strings"
)

func CreateRandomQueue(awsRegion string, prefix string) (string, error) {
	logger := logging.GetLogger("CreateRandomQueue")
	logger.Debugf("Creating randomly named SQS queue with prefix %s", prefix)

	sqsClient, err := CreateSqsClient(awsRegion)
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

	return *queue.QueueUrl, nil;
}

func DeleteQueue(awsRegion string, queueUrl string) (error) {
	logger := logging.GetLogger("DeleteQueue")
	logger.Debugf("Deleting SQS Queue %s", queueUrl)

	sqsClient, err := CreateSqsClient(awsRegion)
	if err != nil {
		return err
	}

	_, err = sqsClient.DeleteQueue(&sqs.DeleteQueueInput{
		QueueUrl:aws.String(queueUrl),
	})

	if err != nil {
		return err
	}
	return nil
}

func DeleteMessageFromQueue(awsRegion string, queueUrl string, receipt string) (error) {
	logger := logging.GetLogger("DeleteMessageFromQueue")
	logger.Debugf("Deleting message from queue %s (%s)", queueUrl, receipt)

	sqsClient, err := CreateSqsClient(awsRegion)
	if err != nil {
		return err
	}

	_, err = sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		ReceiptHandle: &receipt,
		QueueUrl: &queueUrl,
	})
	if err != nil {
		return err
	}

	return nil
}

func SendMessageToQueue(awsRegion string, queueUrl string, message string) (error) {
	logger := logging.GetLogger("SendMessageToQueue")

	sqsClient, err := CreateSqsClient(awsRegion)
	if err != nil {
		return err
	}

	logger.Debugf("Sending message %s to queue %s", message, queueUrl)
	res, err := sqsClient.SendMessage(&sqs.SendMessageInput{
		MessageBody: &message,
		QueueUrl: &queueUrl,
	})

	if err != nil {
		if strings.Contains(err.Error(), "AWS.SimpleQueueService.NonExistentQueue") {
			logger.Warn(fmt.Sprintf("Client has stopped listening on queue %s", queueUrl))
			return nil
		}
		return err
	}
	logger.Debugf("Message id %s sent to queue %s", res.MessageId, queueUrl)

	return nil
}

func CreateSqsClient(awsRegion string) (*sqs.SQS, error) {
	awsConfig, err := CreateAwsConfig(awsRegion)
	if err != nil {
		return nil, err
	}

	return sqs.New(session.New(), awsConfig), nil
}

type QueueMessageResponse struct {
	ReceiptHandle string
	MessageBody   string
	Error         error
}

// Waits to receive a message from on the queueUrl. Since the API only allows us to wait a max 20 seconds for a new
// message to arrive, we must loop TIMEOUT/20 number of times to be able to wait for a total of TIMEOUT seconds
func WaitForQueueMessage(awsRegion string, queueUrl string, timeout int) (QueueMessageResponse) {
	logger := logging.GetLogger("WaitForQueueMessage")

	sqsClient, err := CreateSqsClient(awsRegion)
	if err != nil {
		return QueueMessageResponse{Error:err}
	}

	cycles := timeout;
	cycleLength := 1;

	if timeout >= 20 {
		cycleLength = 20
		cycles = timeout / cycleLength
	}

	for i := 0; i < cycles; i++ {
		logger.Debugf("Waiting for message on %s (%ss)", queueUrl, strconv.Itoa(i * cycleLength))
		result, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl: aws.String(queueUrl),
			AttributeNames: aws.StringSlice([]string{
				"SentTimestamp",
			}),
			MaxNumberOfMessages: aws.Int64(1),
			MessageAttributeNames: aws.StringSlice([]string{
				"All",
			}),
			WaitTimeSeconds: aws.Int64(int64(cycleLength)),
		})

		if err != nil {
			return QueueMessageResponse{Error:err}
		}

		if len(result.Messages) > 0 {
			logger.Debugf("Message %s received on %s", *result.Messages[0].MessageId, queueUrl)
			return QueueMessageResponse{ReceiptHandle:*result.Messages[0].ReceiptHandle, MessageBody:*result.Messages[0].Body}
		}
	}

	return QueueMessageResponse{Error:fmt.Errorf("Failed to receive messages on %s within %s seconds", queueUrl, strconv.Itoa(timeout))}
}
