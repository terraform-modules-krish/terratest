package aws

import (
	"log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/terraform-modules-krish/terratest/util"
	"fmt"
	"time"
	"encoding/base64"
)

// Get the syslog for the Instance with the given ID in the given region. This should be available ~1 minute after an
// Instance boots and is very useful for debugging boot-time issues, such as an error in User Data.
func GetSyslogForInstance(instanceId string, awsRegion string, logger *log.Logger) (string, error) {
	description := fmt.Sprintf("Fetching syslog for Instance %s in %s", instanceId, awsRegion)
	maxRetries := 60
	timeBetweenRetries := 5 * time.Second

	logger.Printf(description)

	ec2Client, err := CreateEC2Client(awsRegion)
	if err != nil {
		return "", err
	}

	input := ec2.GetConsoleOutputInput {
		InstanceId: aws.String(instanceId),
	}

	syslogB64, err := util.DoWithRetry(description, maxRetries, timeBetweenRetries, logger, func() (string, error) {
		out, err := ec2Client.GetConsoleOutput(&input)
		if err != nil {
			return "", err
		}

		syslog := aws.StringValue(out.Output)
		if syslog == "" {
			return "", fmt.Errorf("Syslog is not yet available for instance %s in %s", instanceId, awsRegion)
		}

		return syslog, nil
	})

	if err != nil {
		return "", err
	}

	syslogBytes, err := base64.StdEncoding.DecodeString(syslogB64)
	if err != nil {
		return "", err
	}

	return string(syslogBytes), nil
}

// Get the syslog for each of the Instances in the given ASG in the given region. These logs should be available ~1
// minute after the Instance boots and are very useful for debugging boot-time issues, such as an error in User Data.
// Returns a map of Instance Id -> Syslog for that Instance.
func GetSyslogForInstancesInAsg(asgName string, awsRegion string, logger *log.Logger) (map[string]string, error) {
	logger.Printf("Fetching syslog for each Instance in ASG %s in %s", asgName, awsRegion)

	instanceIds, err := GetEc2InstanceIdsByTag(awsRegion, "aws:autoscaling:groupName", asgName)
	if err != nil {
		return nil, err
	}

	logs := map[string]string{}
	for _, instanceId := range instanceIds {
		syslog, err := GetSyslogForInstance(instanceId, awsRegion, logger)
		if err != nil {
			return nil, err
		}
		logs[instanceId] = syslog
	}

	return logs, nil
}