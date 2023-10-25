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

	isTestDataPresent := IsTestDataPresent(t, "/file/that/does/not/exist", logger)
	assert.False(t, isTestDataPresent, "Expected no test data would be present because no test data file exists.")

	tmpFile, err := ioutil.TempFile("", "save-and-load-test-data")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	expectedData := testData{
		Foo: "foo",
		Bar: true,
		Baz: map[string]interface{}{"abc": "def", "ghi": 1.0, "klm": false},
	}

	isTestDataPresent = IsTestDataPresent(t, tmpFile.Name(), logger)
	assert.False(t, isTestDataPresent, "Expected no test data would be present because file exists but no data has been written yet.")

	SaveTestData(t, tmpFile.Name(), expectedData, logger)

	isTestDataPresent = IsTestDataPresent(t, tmpFile.Name(), logger)
	assert.True(t, isTestDataPresent, "Expected test data would be present because file exists and data has been written to file.")

	actualData := testData{}
	LoadTestData(t, tmpFile.Name(), &actualData, logger)
	assert.Equal(t, expectedData, actualData)

	CleanupTestData(t, tmpFile.Name(), logger)
	assert.False(t, files.FileExists(tmpFile.Name()))
}

func TestIsEmptyJson(t *testing.T) {
	t.Parallel()

	var jsonValue []byte
	var isEmpty bool

	jsonValue = []byte("null")
	isEmpty = isEmptyJson(t, jsonValue)
	assert.True(t, isEmpty, `The JSON literal "null" should be treated as an empty value.`)

	jsonValue = []byte("false")
	isEmpty = isEmptyJson(t, jsonValue)
	assert.True(t, isEmpty, `The JSON literal "false" should be treated as an empty value.`)

	jsonValue = []byte("true")
	isEmpty = isEmptyJson(t, jsonValue)
	assert.False(t, isEmpty, `The JSON literal "true" should be treated as a non-empty value.`)

	jsonValue = []byte("0")
	isEmpty = isEmptyJson(t, jsonValue)
	assert.True(t, isEmpty, `The JSON literal "0" should be treated as an empty value.`)

	jsonValue = []byte("1")
	isEmpty = isEmptyJson(t, jsonValue)
	assert.False(t, isEmpty, `The JSON literal "1" should be treated as a non-empty value.`)

	jsonValue = []byte("{}")
	isEmpty = isEmptyJson(t, jsonValue)
	assert.True(t, isEmpty, `The JSON value "{}" should be treated as an empty value.`)

	jsonValue = []byte(`{ "key": "val" }`)
	isEmpty = isEmptyJson(t, jsonValue)
	assert.False(t, isEmpty, `The JSON value { "key": "val" } should be treated as a non-empty value.`)

	jsonValue = []byte(`[]`)
	isEmpty = isEmptyJson(t, jsonValue)
	assert.True(t, isEmpty, `The JSON value "[]" should be treated as an empty value.`)

	jsonValue = []byte(`[{ "key": "val" }]`)
	isEmpty = isEmptyJson(t, jsonValue)
	assert.False(t, isEmpty, `The JSON value [{ "key": "val" }] should be treated as a non-empty value.`)
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
	assert.False(t, files.FileExists(formatNamedTestDataPath(tmpFolder, "AMI")))
}

func TestSaveAndLoadNamedTestData(t *testing.T) {
	t.Parallel()

	logger := terralog.NewLogger("TestSaveAndLoadAmiId")

	tmpFolder, err := ioutil.TempDir("", "save-and-load-ami-id")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	name1 := "test-ami"
	expectedData1 := "ami-abcd1234"

	name2 := "test-ami2"
	expectedData2 := "ami-xyz98765"

	SaveString(t, tmpFolder, name1, expectedData1, logger)
	SaveString(t, tmpFolder, name2, expectedData2, logger)

	actualData1 := LoadString(t, tmpFolder, name1, logger)
	actualData2 := LoadString(t, tmpFolder, name2, logger)

	assert.Equal(t, expectedData1, actualData1)
	assert.Equal(t, expectedData2, actualData2)

	CleanupNamedTestData(t, tmpFolder, name1, logger)
	CleanupNamedTestData(t, tmpFolder, name2, logger)

	assert.False(t, files.FileExists(formatNamedTestDataPath(tmpFolder, name1)))
	assert.False(t, files.FileExists(formatNamedTestDataPath(tmpFolder, name2)))
}

func TestSaveDuplicateTestData(t *testing.T) {
	t.Parallel()

	logger := terralog.NewLogger("TestSaveAndLoadAmiId")

	tmpFolder, err := ioutil.TempDir("", "save-and-load-ami-id")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	name := "hello-world"
	val1 := "hello world"
	val2 := "buenos dias, mundo"

	SaveString(t, tmpFolder, name, val1, logger)
	SaveString(t, tmpFolder, name, val2, logger)

	actualVal := LoadString(t, tmpFolder, name, logger)

	assert.Equal(t, val2, actualVal, "Actual test data should use overwritten values")
}