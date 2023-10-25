// Integration tests that test cross-package functionality in AWS.
package terratest

import (
	"testing"
	"fmt"
	"github.com/terraform-modules-krish/terratest/util"
)

func TestCreateRandomResourceCollectionOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		forbiddenRegions []string
		approvedRegions []string
		validRegions []string
	}{
		{
			[]string{ "ap-south-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ap-northeast-2", "ca-central-1", "eu-central-1", "eu-west-1", "eu-west-2", "sa-east-1", "us-west-1", "us-west-2", },
			[]string{},
			[]string{ "us-east-1", "us-east-2", },
		},
		{
			[]string{ "ap-south-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ap-northeast-2", "ca-central-1", "eu-central-1", "eu-west-1", "eu-west-2", "sa-east-1", "us-west-1", "us-west-2", },
			[]string{ "us-east-1", "us-east-2", },
			[]string{ "us-east-1", "us-east-2", },
		},
		{
			[]string{ "ap-south-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ap-northeast-2", "ca-central-1", "eu-central-1", "eu-west-1", "eu-west-2", "sa-east-1", "us-west-1", "us-west-2", },
			[]string{ "us-east-1", },
			[]string{ "us-east-1", },
		},
		{
			[]string{},
			[]string{ "us-west-1", },
			[]string{ "us-west-1", },
		},
		{
			[]string{},
			[]string{ "ap-south-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ap-northeast-2", "ca-central-1", "eu-central-1", "eu-west-1", "eu-west-2", "sa-east-1", "us-east-1", "us-east-2", },
			[]string{ "ap-south-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ap-northeast-2", "ca-central-1", "eu-central-1", "eu-west-1", "eu-west-2", "sa-east-1", "us-east-1", "us-east-2", },
		},
	}

	for _, testCase := range testCases {
		ro := NewRandomResourceCollectionOptions()
		ro.ForbiddenRegions = testCase.forbiddenRegions
		ro.ApprovedRegions = testCase.approvedRegions

		rand, err := CreateRandomResourceCollection(ro)
		defer rand.DestroyResources()
		if err != nil {
			t.Fatalf("Failed to create RandomResourceCollection: %s", err.Error())
		}

		if ! util.ListContains(rand.AwsRegion, testCase.validRegions) {
			t.Fatalf("An invalid region was selected. The only valid regions were %v, but selected region was %s", testCase.validRegions, rand.AwsRegion)
		}
	}
}

func TestFetchAwsAvailabilityZones(t *testing.T) {
	t.Parallel()

	ro := NewRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Fatalf("Failed to create RandomResourceCollection: %s", err.Error())
	}

	// Manually set the AWS Region to us-west-2 for testing purposes
	rand.AwsRegion = "us-west-2"
	actual := rand.FetchAwsAvailabilityZones()
	expected := []string{"us-west-2a","us-west-2b","us-west-2c"}

	for index,_ := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("Expected: %s, but received %s", expected[index], actual[index])
		}
	}
}

func TestGetRandomPrivateCidrBlock(t *testing.T) {
	t.Parallel()

	ro := NewRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Fatalf("Failed to create RandomResourceCollection: %s", err.Error())
	}

	actual := rand.GetRandomPrivateCidrBlock(18)
	actualPrefix := string(actual[len(actual)-3:])
	expPrefix := "/18"

	if actualPrefix != expPrefix {
		t.Fatalf("Expected: %s, but received: %s", expPrefix, actualPrefix)
	}
}

func TestAllParametersSet(t *testing.T) {
	t.Parallel()

	ro := NewRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Fatalf("Failed to create RandomResourceCollection: %s", err.Error())
	}

	if len(rand.AccountId) == 0 {
		t.Fatalf("CreateRandomResourceCollection has an empty AccountId: %+v", rand)
	}

	if len(rand.AmiId) == 0 {
		t.Fatalf("CreateRandomResourceCollection has an empty AMI ID: %+v", rand)
	}

	if len(rand.AwsRegion) == 0 {
		t.Fatalf("CreateRandomResourceCollection has an empty region: %+v", rand)
	}

	if len(rand.UniqueId) == 0 {
		t.Fatalf("CreateRandomResourceCollection has an empty Unique Id: %+v", rand)
	}

	if rand.KeyPair == nil {
		t.Fatalf("CreateRandomResourceCollection has a nil Key Pair: %+v", rand)
	}

	if len(rand.SnsTopicArn) == 0 {
		t.Fatalf("CreateRandomResourceCollection has an empty SnsTopicArn: %+v", rand)
	}

	fmt.Printf("%+v", rand)
}

func TestGetDefaultVpc(t *testing.T) {
	t.Parallel()

	ro := NewRandomResourceCollectionOptions()
	rand, err := CreateRandomResourceCollection(ro)
	defer rand.DestroyResources()
	if err != nil {
		t.Fatalf("Failed to create RandomResourceCollection: %s", err.Error())
	}

	vpc, err := rand.GetDefaultVpc()
	if err != nil {
		t.Fatalf("Failed to get random VPC: %s", err.Error())
	}

	if vpc.Id == "" {
		t.Fatalf("GetRandomVpc returned a VPC without an ID: %s", vpc)
	}

	if vpc.Name == "" {
		t.Fatalf("GetRandomVpc returned a VPC without a name: %s", vpc)
	}

	if len(vpc.Subnets) == 0 {
		t.Fatalf("GetRandomVpc returned a VPC with no subnets: %s", vpc)
	}
}