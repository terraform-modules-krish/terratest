package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"fmt"
	"time"
	"encoding/base64"
	"testing"
	"github.com/terraform-modules-krish/terratest/modules/logger"
	"github.com/terraform-modules-krish/terratest/modules/retry"
)

// Get the syslog for the Instance with the given ID in the given region. This should be available ~1 minute after an
// Instance boots and is very useful for debugging boot-time issues, such as an error in User Data.
func GetSyslogForInstance(t *testing.T, instanceId string, awsRegion string) string {
	out, err := GetSyslogForInstanceE(t, instanceId, awsRegion)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Get the syslog for the Instance with the given ID in the given region. This should be available ~1 minute after an
// Instance boots and is very useful for debugging boot-time issues, such as an error in User Data.
func GetSyslogForInstanceE(t *testing.T, instanceId string, region string) (string, error) {
	description := fmt.Sprintf("Fetching syslog for Instance %s in %s", instanceId, region)
	maxRetries := 60
	timeBetweenRetries := 5 * time.Second

	logger.Log(t, description)

	client, err := NewEc2ClientE(t, region)
	if err != nil {
		return "", err
	}

	input := ec2.GetConsoleOutputInput {
		InstanceId: aws.String(instanceId),
	}

	syslogB64, err := retry.DoWithRetryE(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		out, err := client.GetConsoleOutput(&input)
		if err != nil {
			return "", err
		}

		syslog := aws.StringValue(out.Output)
		if syslog == "" {
			return "", fmt.Errorf("Syslog is not yet available for instance %s in %s", instanceId, region)
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
func GetSyslogForInstancesInAsg(t *testing.T, asgName string, awsRegion string) map[string]string {
	out, err := GetSyslogForInstancesInAsgE(t, asgName, awsRegion)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Get the syslog for each of the Instances in the given ASG in the given region. These logs should be available ~1
// minute after the Instance boots and are very useful for debugging boot-time issues, such as an error in User Data.
// Returns a map of Instance Id -> Syslog for that Instance.
func GetSyslogForInstancesInAsgE(t *testing.T, asgName string, awsRegion string) (map[string]string, error) {
	logger.Logf(t, "Fetching syslog for each Instance in ASG %s in %s", asgName, awsRegion)

	instanceIds, err := GetEc2InstanceIdsByTagE(t, awsRegion, "aws:autoscaling:groupName", asgName)
	if err != nil {
		return nil, err
	}

	logs := map[string]string{}
	for _, instanceId := range instanceIds {
		syslog, err := GetSyslogForInstanceE(t, instanceId, awsRegion)
		if err != nil {
			return nil, err
		}
		logs[instanceId] = syslog
	}

	return logs, nil
}