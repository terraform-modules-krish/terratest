package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"testing"
	"github.com/terraform-modules-krish/terratest/modules/logger"
)

// Get the public IP address of the given EC2 Instance in the given region
func GetPublicIpOfEc2Instance(t *testing.T, instanceId string, awsRegion string) string {
	ip, err := GetPublicIpOfEc2InstanceE(t, instanceId, awsRegion)
	if err != nil {
		t.Fatal(err)
	}
	return ip
}

// Get the public IP address of the given EC2 Instance in the given region
func GetPublicIpOfEc2InstanceE(t *testing.T, instanceId string, awsRegion string) (string, error) {
	ec2Client := NewEc2Client(t, awsRegion)

	input := ec2.DescribeInstancesInput{InstanceIds: []*string{aws.String(instanceId)}}
	output, err := ec2Client.DescribeInstances(&input)
	if err != nil {
		return "", err
	}

	for _, reserveration := range output.Reservations {
		for _, instance := range reserveration.Instances {
			return aws.StringValue(instance.PublicIpAddress), nil
		}
	}

	return "", IpForEc2InstanceNotFound{InstanceId: instanceId, AwsRegion: awsRegion}
}

// Return all the IDs of EC2 instances in the given region with the given tag
func GetEc2InstanceIdsByTag(t *testing.T, region string, tagName string, tagValue string) []string {
	out, err := GetEc2InstanceIdsByTagE(t, region, tagName, tagValue)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Return all the IDs of EC2 instances in the given region with the given tag
func GetEc2InstanceIdsByTagE(t *testing.T, region string, tagName string, tagValue string) ([]string, error) {
	client, err := NewEc2ClientE(t, region)
	if err != nil {
		return nil, err
	}

	tagFilter := &ec2.Filter{
		Name:   aws.String(fmt.Sprintf("tag:%s", tagName)),
		Values: []*string{aws.String(tagValue)},
	}
	output, err := client.DescribeInstances(&ec2.DescribeInstancesInput{Filters: []*ec2.Filter{tagFilter}})
	if err != nil {
		return nil, err
	}

	instanceIds := []string{}

	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			instanceIds = append(instanceIds, *instance.InstanceId)
		}
	}

	return instanceIds, err
}

// Return all the tags for the given EC2 Instance
func GetTagsForEc2Instance(t *testing.T, region string, instanceId string) map[string]string {
	tags, err := GetTagsForEc2InstanceE(t, region, instanceId)
	if err != nil {
		t.Fatal(err)
	}
	return tags
}

// Return all the tags for the given EC2 Instance
func GetTagsForEc2InstanceE(t *testing.T, region string, instanceId string) (map[string]string, error) {
	client, err := NewEc2ClientE(t, region)
	if err != nil {
		return nil, err
	}

	input := ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("resource-type"),
				Values: aws.StringSlice([]string{"instance"}),
			},
			{
				Name:   aws.String("resource-id"),
				Values: aws.StringSlice([]string{instanceId}),
			},
		},
	}

	out, err := client.DescribeTags(&input)
	if err != nil {
		return nil, err
	}

	tags := map[string]string{}

	for _, tag := range out.Tags {
		tags[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
	}

	return tags, nil
}

// Delete the given AMI in the given region
func DeleteAmi(t *testing.T, region string, imageId string) {
	err := DeleteAmiE(t, region, imageId)
	if err != nil {
		t.Fatal(err)
	}
}

// Delete the given AMI in the given region
func DeleteAmiE(t *testing.T, region string, imageId string) error {
	logger.Logf(t, "Deregistering AMI %s", imageId)

	client, err := NewEc2ClientE(t, region)
	if err != nil {
		return err
	}

	_, err = client.DeregisterImage(&ec2.DeregisterImageInput{ImageId: aws.String(imageId)})
	return err
}

// Terminate the EC2 instance with the given ID in the given region
func TerminateInstance(t *testing.T, region string, instanceId string) {
	err := TerminateInstanceE(t, region, instanceId)
	if err != nil {
		t.Fatal(err)
	}
}

// Terminate the EC2 instance with the given ID in the given region
func TerminateInstanceE(t *testing.T, region string, instanceId string) error {
	logger.Logf(t, "Terminating Instance %s", instanceId)

	client, err := NewEc2ClientE(t, region)
	if err != nil {
		return err
	}

	_, err = client.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceId),
		},
	})

	return err
}

// Create an EC2 client
func NewEc2Client(t *testing.T, region string) *ec2.EC2 {
	client, err := NewEc2ClientE(t, region)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// Create an EC2 client
func NewEc2ClientE(t *testing.T, region string) (*ec2.EC2, error) {
	sess, err := NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}

	return ec2.New(sess), nil
}

type IpForEc2InstanceNotFound struct {
	InstanceId string
	AwsRegion  string
}

func (err IpForEc2InstanceNotFound) Error() string {
	return fmt.Sprintf("Could not find a public IP address for EC2 Instance %s in %s", err.InstanceId, err.AwsRegion)
}
