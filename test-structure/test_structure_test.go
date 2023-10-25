package test_structure

import (
	"testing"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	terralog "github.com/terraform-modules-krish/terratest/log"
	"github.com/terraform-modules-krish/terratest/files"
	"github.com/terraform-modules-krish/terratest"
)

type testData struct {
	Foo string
	Bar bool
	Baz map[string]interface{}
}

func TestSaveAndLoadTestData(t *testing.T) {
	t.Parallel()

	logger := terralog.NewLogger("TestSaveAndLoadTestData")

	tmpFile, err := ioutil.TempFile("", "save-and-load-test-data")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	expectedData := testData{
		Foo: "foo",
		Bar: true,
		Baz: map[string]interface{}{"abc": "def", "ghi": 1.0, "klm": false},
	}
	SaveTestData(t, tmpFile.Name(), expectedData, logger)

	actualData := testData{}
	LoadTestData(t, tmpFile.Name(), &actualData, logger)
	assert.Equal(t, expectedData, actualData)

	CleanupTestData(t, tmpFile.Name(), logger)
	assert.False(t, files.FileExists(tmpFile.Name()))
}

func TestSaveAndLoadTerratestOptions(t *testing.T) {
	t.Parallel()

	logger := terralog.NewLogger("TestSaveAndLoadTerratestOptions")

	tmpFolder, err := ioutil.TempDir("", "save-and-load-terratest-options")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	expectedData := &terratest.TerratestOptions{
		UniqueId: "foo",
		TestName: "bar",
		TemplatePath: "/abc/def/ghi",
		Vars: map[string]interface{}{},
	}
	SaveTerratestOptions(t, tmpFolder, expectedData, logger)

	actualData := LoadTerratestOptions(t, tmpFolder, logger)
	assert.Equal(t, expectedData, actualData)

	CleanupTerratestOptions(t, tmpFolder, logger)
	assert.False(t, files.FileExists(formatTerratestOptionsPath(tmpFolder)))
}

func TestSaveAndLoadRandomResourceCollection(t *testing.T) {
	t.Parallel()

	logger := terralog.NewLogger("TestSaveAndLoadRandomResourceCollection")

	tmpFolder, err := ioutil.TempDir("", "save-and-load-random-resource-collection")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	expectedData := &terratest.RandomResourceCollection{
		AwsRegion: "us-east-1",
		UniqueId: "foo",
		AccountId: "1234567890",
		KeyPair: &terratest.Ec2Keypair{
			Name: "foo",
			PublicKey: "fake-public-key",
			PrivateKey: "fake-private-key",
		},
		AmiId: "ami-abcd1234",
	}
	SaveRandomResourceCollection(t, tmpFolder, expectedData, logger)

	actualData := LoadRandomResourceCollection(t, tmpFolder, logger)
	assert.Equal(t, expectedData, actualData)

	CleanupRandomResourceCollection(t, tmpFolder, logger)
	assert.False(t, files.FileExists(formatRandomResourceCollectionPath(tmpFolder)))
}

func TestSaveAndLoadAmiId(t *testing.T) {
	t.Parallel()

	logger := terralog.NewLogger("TestSaveAndLoadAmiId")

	tmpFolder, err := ioutil.TempDir("", "save-and-load-ami-id")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	expectedData := "ami-abcd1234"
	SaveAmiId(t, tmpFolder, expectedData, logger)

	actualData := LoadAmiId(t, tmpFolder, logger)
	assert.Equal(t, expectedData, actualData)

	CleanupAmiId(t, tmpFolder, logger)
	assert.False(t, files.FileExists(formatAmiIdPath(tmpFolder)))
}
