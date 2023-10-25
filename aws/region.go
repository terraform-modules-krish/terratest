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
func GetRandomRegion(approvedRegions, forbiddenRegions []string) string {
	log := log.NewLogger("GetRandomRegion")

	allRegions := []string{
		"ap-south-1",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"ap-northeast-2",
		"ca-central-1",
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		"sa-east-1",
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
	}

	var randomIndex int
	selectedRegionIsValid := false

	// Select a random region
	// Make sure that it's both an approved region (if ro.ApprovedRegions is non-empty) and not in ro.ForbiddenRegions or .
	for i := 0; i < 1000 && !selectedRegionIsValid; i++ {
		randomIndex = util.Random(0,len(allRegions))
		selectedRegion := allRegions[randomIndex]

		regionIsApproved := false
		regionIsForbidden := false

		if len(approvedRegions) == 0 {
			regionIsApproved = true
		} else if util.ListContains(selectedRegion, approvedRegions) {
			regionIsApproved = true
		}

		for _, forbiddenRegion := range GetGloballyForbiddenRegions() {
			if forbiddenRegion == selectedRegion {
				regionIsForbidden = true
			}
		}

		for _, forbiddenRegion := range forbiddenRegions {
			if forbiddenRegion == selectedRegion {
				regionIsForbidden = true
			}
		}

		if regionIsApproved && !regionIsForbidden {
			selectedRegionIsValid = true
		}
	}

	if ! selectedRegionIsValid {
		log.Println("WARNING: Attempted to select an AWS region 1,000 times and still couldn't find a valid region.")
		return "<GetRandomRegions-could-not-select-a-region>"
	} else {
		return allRegions[randomIndex]
	}
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

