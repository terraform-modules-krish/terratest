// Integration tests that validate S3-related code in AWS.
package aws

import (
	"strings"
	"testing"

	"github.com/terraform-modules-krish/terratest/log"
	"github.com/terraform-modules-krish/terratest/util"
)

func TestCreateAndDestroyS3Bucket(t *testing.T) {
	t.Parallel()

	logger := log.NewLogger("TestCreateAndDestroyS3Bucket")

	// SETUP
	region := GetRandomRegion(nil, nil)
	id := util.UniqueId()
	logger.Printf("Random values selected. Region = %s, Id = %s\n", region, id)

	// TEST
	s3BucketName := "gruntwork-terratest-" + strings.ToLower(id)

	CreateS3Bucket(region, s3BucketName)
	DeleteS3Bucket(region, s3BucketName)
}

func TestAssertS3BucketExistsNoFalseNegative(t *testing.T) {
	t.Parallel()

	logger := log.NewLogger("TestAssertS3BucketExists")

	// SETUP
	region := GetRandomRegion(nil, nil)
	s3BucketName := "gruntwork-terratest-" + strings.ToLower(util.UniqueId())
	logger.Printf("Random values selected. Region = %s, s3BucketName = %s\n", region, s3BucketName)

	CreateS3Bucket(region, s3BucketName)

	// TEST
	err := AssertS3BucketExists(region, s3BucketName)
	if err != nil {
		t.Fatalf("Function claimed that S3 Bucket '%s' does not exist, but in fact it does.", s3BucketName)
	}

	// TEARDOWN
	DeleteS3Bucket(region, s3BucketName)
}

func TestAssertS3BucketExistsNoFalsePositive(t *testing.T) {
	t.Parallel()

	logger := log.NewLogger("TestAssertS3BucketExists")

	// SETUP
	region := GetRandomRegion(nil, nil)
	s3BucketName := "gruntwork-terratest-" + strings.ToLower(util.UniqueId())
	logger.Printf("Random values selected. Region = %s, s3BucketName = %s\n", region, s3BucketName)

	// We elect not to create the S3 bucket to confirm that our function correctly reports it doesn't exist.
	//aws.CreateS3Bucket(region, s3BucketName)

	// TEST
	err := AssertS3BucketExists(region, s3BucketName)
	if err == nil {
		t.Fatalf("Function claimed that S3 Bucket '%s' exists, but in fact it does not.", s3BucketName)
	}
}