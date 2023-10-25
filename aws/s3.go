package aws

import (
	"os"
	"strings"

	"github.com/terraform-modules-krish/terratest/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/session"
"errors"
	"fmt"
	gruntwork-cli "github.com/terraform-modules-krish/go-commons/logging"
	"bytes"
)

func CreateS3Client(awsRegion string) (*s3.S3, error) {
	awsConfig, err := CreateAwsConfig(awsRegion)
	if err != nil {
		return nil, err
	}

	return s3.New(session.New(), awsConfig), nil
}

func FindS3BucketWithTag(awsRegion string, key string, value string) (string, error) {
	logger := logging.GetLogger("FindS3BucketWithTag")

	s3Client, err := CreateS3Client(awsRegion)
	if err != nil {
		return "", err
	}

	resp, err := s3Client.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return "", err
	}

	for _, bucket := range resp.Buckets {
		tagResponse, err := s3Client.GetBucketTagging(&s3.GetBucketTaggingInput{Bucket: bucket.Name})
		if err != nil {

			if !strings.Contains(err.Error(), "AuthorizationHeaderMalformed") &&
				!strings.Contains(err.Error(), "BucketRegionError") &&
				!strings.Contains(err.Error(), "NoSuchTagSet") {
				return "", err
			}

		}

		for _, tag := range tagResponse.TagSet {
			if *tag.Key == key && *tag.Value == value {
				logger.Debugf("Found S3 bucket %s with %s=%s", *bucket.Name, key, value)
				return *bucket.Name, nil
			}
		}
	}

	return "", nil
}

func GetS3ObjectContents(awsRegion string, bucket string, key string) (string, error) {
	logger := logging.GetLogger("GetS3ObjectContents")

	s3Client, err := CreateS3Client(awsRegion)
	if err != nil {
		return "", err
	}

	res, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: &bucket,
		Key: &key,
	})

	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return "", err
	}

	contents := buf.String()
	logger.Debugf("Read contents from s3://%s/%s", bucket, key)

	return contents, nil
}

// Create an S3 bucket.
func CreateS3Bucket(region string, name string) {
	log := log.NewLogger("CreateS3Bucket")

	svc := s3.New(session.New(), aws.NewConfig().WithRegion(region))
	_, err := svc.Config.Credentials.Get()
	if err != nil {
		log.Fatalf("Failed to open S3 session: %s\n", err.Error())
	}

	params := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}
	_, err = svc.CreateBucket(params)
	if err != nil {
		log.Printf("Failed to create S3 bucket: %s", err.Error())
		os.Exit(1)
	}
}

// Destroy an S3 bucket.
func DeleteS3Bucket(region string, name string) {
	log := log.NewLogger("DestroyS3Bucket")

	svc := s3.New(session.New(), aws.NewConfig().WithRegion(region))
	_, err := svc.Config.Credentials.Get()
	if err != nil {
		log.Fatalf("Failed to open S3 session: %s\n", err.Error())
	}

	params := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}
	_, err = svc.DeleteBucket(params)
	if err != nil {
		log.Printf("Failed to delete S3 bucket: %s", err.Error())
		os.Exit(1)
	}
}

// Returns true if the given S3 bucket exists.
func AssertS3BucketExists(region string, name string) error {
	log := log.NewLogger("AssertS3BucketExists")

	svc := s3.New(session.New(), aws.NewConfig().WithRegion(region))
	_, err := svc.Config.Credentials.Get()
	if err != nil {
		log.Printf("Failed to open S3 session: %s\n", err.Error())
	}

	params := &s3.HeadBucketInput{
		Bucket: aws.String(name),
	}
	_, err = svc.HeadBucket(params)

	if err != nil {
		// We expect a missing bucket to return this error code.  Otherwise, fail because we can't be sure what
		// the AWS response means.
		if ! strings.Contains(err.Error(), "status code: 404") {
			log.Printf("Failed to assert whether bucket exists: %s", err.Error())
			os.Exit(1)
		} else {
			return errors.New(fmt.Sprintf("Assertion that S3 Bucket '%s' exists failed. That bucket does not exist.", name))
		}
	}

	return nil
}