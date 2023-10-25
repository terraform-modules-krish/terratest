package resources

import (
	"testing"
	"github.com/terraform-modules-krish/terratest"
	"github.com/terraform-modules-krish/terratest/terraform"
	terraws "github.com/terraform-modules-krish/terratest/aws"
)

func CreateBaseTerratestOptions(t *testing.T, testName string, templatePath string, resourceCollection *terratest.RandomResourceCollection) *terratest.TerratestOptions {
	terratestOptions := terratest.NewTerratestOptions()

	terratestOptions.UniqueId = resourceCollection.UniqueId
	terratestOptions.TemplatePath = templatePath
	terratestOptions.TestName = testName

	vpc, err := resourceCollection.GetDefaultVpc()
	if err != nil {
		t.Fatalf("Failed to get default VPC: %s", err.Error())
	}

	terratestOptions.Vars = map[string]interface{} {
		"aws_region": resourceCollection.AwsRegion,
		"aws_account_id": resourceCollection.AccountId,
		"vpc_id": vpc.Id,
		"subnet_ids": getSubnetIds(vpc.Subnets),
		"name": testName + resourceCollection.UniqueId,
	}

	terratestOptions.RetryableTerraformErrors = map[string]string {
		terraform.TF_ERROR_DIFFS_DIDNT_MATCH_DURING_APPLY: terraform.TF_ERROR_DIFFS_DIDNT_MATCH_DURING_APPLY_MSG,
		terraform.TF_ERROR_FINDING_MATCHING_ROUTE_FOR_ROUTE_TABLE: terraform.TF_ERROR_FINDING_MATCHING_ROUTE_FOR_ROUTE_TABLE_MSG,
		terraform.TF_ERROR_DOES_NOT_HAVE_ATTRIBUTE_ID_FOR_VARIABLE: terraform.TF_ERROR_DOES_NOT_HAVE_ATTRIBUTE_ID_FOR_VARIABLE_MSG,
		terraform.TF_ERROR_INVALID_ROUTE_TABLE_ID: terraform.TF_ERROR_INVALID_ROUTE_TABLE_ID_MSG,
		terraform.TF_ERROR_EXPECTED_TO_FIND_ONE_NETWORK_ACL: terraform.TF_ERROR_EXPECTED_TO_FIND_ONE_NETWORK_ACL_MSG,
		terraform.TF_ERROR_INVALID_SUBNET_ID: terraform.TF_ERROR_INVALID_SUBNET_ID_MSG,
		terraform.TF_ERROR_FINDING_ROUTE_AFTER_CREATING: terraform.TF_ERROR_FINDING_ROUTE_AFTER_CREATING_MSG,
	}

	return terratestOptions
}


func getSubnetIds(subnets []terraws.Subnet) []string {
	subnetIds := []string{}

	for _, subnet := range subnets {
		subnetIds = append(subnetIds, subnet.Id)
	}

	return subnetIds
}

