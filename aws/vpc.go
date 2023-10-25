package aws
import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"errors"
	"github.com/terraform-modules-krish/terratest/util"
)

var VpcIdFilterName = "vpc-id"

type Vpc struct {
	Id         string 		// The ID of the VPC
	Name       string		// The name of the VPC
	Subnets    []ec2.Subnet	// A list of subnets in the VPC
}

func GetRandomVpc(awsRegion string) (Vpc, error) {
	vpc := Vpc{}

	svc := ec2.New(session.New(), aws.NewConfig().WithRegion(awsRegion))
	vpcs, err := svc.DescribeVpcs(&ec2.DescribeVpcsInput{})
	if err != nil {
		return vpc, err
	}

	numVpcs := len(vpcs.Vpcs)
	if numVpcs == 0 {
		return vpc, errors.New("No VPCs found in region " + awsRegion)
	}

	randomIndex := util.Random(0, numVpcs)
	randomVpc := vpcs.Vpcs[randomIndex]

	vpc.Id = *randomVpc.VpcId
	vpc.Name = FindVpcName(randomVpc)

	vpc.Subnets, err = GetSubnetsForVpc(vpc.Id, awsRegion)
	return vpc, err
}

func FindVpcName(vpc *ec2.Vpc) string {
	for _, tag := range vpc.Tags {
		if *tag.Key == "Name" {
			return *tag.Value
		}
	}

	if *vpc.IsDefault {
		return "Default"
	}

	return ""
}

func GetSubnetsForVpc(vpcId string, awsRegion string) ([]ec2.Subnet, error) {
	subnets := []ec2.Subnet{}

	svc := ec2.New(session.New(), aws.NewConfig().WithRegion(awsRegion))

	vpcIdFilter := ec2.Filter{Name: &VpcIdFilterName, Values: []*string{&vpcId}}
	subnetOutput, err := svc.DescribeSubnets(&ec2.DescribeSubnetsInput{Filters: []*ec2.Filter{&vpcIdFilter}})
	if err != nil {
		return subnets, err
	}

	for _, subnet := range subnetOutput.Subnets {
		subnets = append(subnets, *subnet)
	}
	return subnets, nil
}