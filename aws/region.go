package aws

import (
	"github.com/terraform-modules-krish/terratest/util"
	"github.com/terraform-modules-krish/terratest/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func GetGloballyForbiddenRegions() []string {
	return []string{
		"us-west-2",		// Josh is using this region for his personal projects
		"ap-northeast-2",	// This region seems to be running out of t2.micro instances with gp2 volumes
	}
}

// Get a randomly chosen AWS region that's not in the forbiddenRegions list
func GetRandomRegion(forbiddenRegions []string) string {

	allRegions := []string{
		"us-east-1",
		"us-west-1",
		"us-west-2",
		"eu-west-1",
		"eu-central-1",
		"ap-northeast-1",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"sa-east-1",
	}

	// Select a random region
	// If our randomIndex gave us a region that's forbidden, keep iterating until we get a valid one.
	var randomIndex int
	randomIndexIsValid := false

	for !randomIndexIsValid {
		randomIndex = util.Random(0,len(allRegions))
		randomIndexIsValid = true

		// The ... allows append to be used to concatenate two slices
		for _, forbiddenRegion := range append(GetGloballyForbiddenRegions(), forbiddenRegions...) {
			if forbiddenRegion == allRegions[randomIndex] {
				randomIndexIsValid = false
			}
		}
	}

	return allRegions[randomIndex]
}

// Get the Availability Zones for a given AWS region. Note that for certain regions (e.g. us-east-1), different AWS
// accounts have access to different availability zones.
func GetAvailabilityZones(region string) []string {
	log := log.NewLogger("GetAvailabilityZones")

	svc := ec2.New(session.New(), aws.NewConfig().WithRegion(region))
	_, err := svc.Config.Credentials.Get()
	if err != nil {
		log.Fatalf("Failed to open EC2 session: %s\n", err.Error())
	}

	params := &ec2.DescribeAvailabilityZonesInput{
		DryRun: aws.Bool(false),
	}
	resp, err := svc.DescribeAvailabilityZones(params)
	if err != nil {
		log.Fatalf("Failed to fetch AWS Availability Zones: %s\n", err.Error())
	}

	var azs []string
	for _, availabilityZone := range resp.AvailabilityZones {
		azs = append(azs, *availabilityZone.ZoneName)
	}

	return azs
}

