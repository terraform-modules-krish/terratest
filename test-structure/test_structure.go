package test_structure

import (
	"os"
	"testing"
	"encoding/json"
	"io/ioutil"
	"github.com/terraform-modules-krish/terratest"
	"path/filepath"
	"log"
	"github.com/terraform-modules-krish/terratest/files"
	"fmt"
	"strings"
)

const SKIP_STAGE_ENV_VAR_PREFIX = "SKIP_"

// Execute the given test stage (e.g., setup, teardown, validation) if an environment variable of the name
// `SKIP_<stageName>` (e.g., SKIP_teardown) is not set.
func RunTestStage(stageName string, logger *log.Logger, stage func()) {
	envVarName := fmt.Sprintf("%s%s", SKIP_STAGE_ENV_VAR_PREFIX, stageName)
	if os.Getenv(envVarName) == "" {
		logger.Printf("The '%s' environment variable is not set, so executing stage '%s'.", envVarName, stageName)
		stage()
	} else {
		logger.Printf("The '%s' environment variable is set, so skipping stage '%s'.", envVarName, stageName)
	}
}

// Returns true if an environment variable is set instructing Terratest to skip a test stage. This can be an easy way
// to tell if the tests are running in a local dev environment vs a CI server.
func SkipStageEnvVarSet() bool {
	for _, environmentVariable := range os.Environ() {
		if strings.HasPrefix(environmentVariable, SKIP_STAGE_ENV_VAR_PREFIX) {
			return true
		}
	}

	return false
}

// Serialize and save TerratestOptions into the given folder. This allows you to create TerratestOptions during setup
// and to reuse that TerratestOptions later during validation and teardown.
func SaveTerratestOptions(t *testing.T, testFolder string, terratestOptions *terratest.TerratestOptions, logger *log.Logger) {
	SaveTestData(t, formatTerratestOptionsPath(testFolder), terratestOptions, logger)
}

// Load and unserialize TerratestOptions from the given folder. This allows you to reuse a TerratestOptions that was
// created during an earlier setup step in later validation and teardown steps.
func LoadTerratestOptions(t *testing.T, testFolder string, logger *log.Logger) *terratest.TerratestOptions {
	var terratestOptions terratest.TerratestOptions
	LoadTestData(t, formatTerratestOptionsPath(testFolder), &terratestOptions, logger)
	return &terratestOptions
}

// Clean up the files used to store TerratestOptions between test stages
func CleanupTerratestOptions(t *testing.T, testFolder string, logger *log.Logger) {
	CleanupTestData(t, formatTerratestOptionsPath(testFolder), logger)
}

// Format a path to save TerratestOptions in the given folder
func formatTerratestOptionsPath(testFolder string) string {
	return FormatTestDataPath(testFolder, "TerratestOptions.json")
}

// Serialize and save RandomResourceCollection into the given folder. This allows you to create RandomResourceCollection
// during setup and to reuse that RandomResourceCollection later during validation and teardown.
func SaveRandomResourceCollection(t *testing.T, testFolder string, resourceCollection *terratest.RandomResourceCollection, logger *log.Logger) {
	SaveTestData(t, formatRandomResourceCollectionPath(testFolder), resourceCollection, logger)
}

// Load and unserialize RandomResourceCollection from the given folder. This allows you to reuse a RandomResourceCollection
// that was created during an earlier setup step in later validation and teardown steps.
func LoadRandomResourceCollection(t *testing.T, testFolder string, logger *log.Logger) *terratest.RandomResourceCollection {
	var resourceCollection terratest.RandomResourceCollection
	LoadTestData(t, formatRandomResourceCollectionPath(testFolder), &resourceCollection, logger)
	return &resourceCollection
}

// Clean up the files used to store RandomResourceCollection between test stages
func CleanupRandomResourceCollection(t *testing.T, testFolder string, logger *log.Logger) {
	CleanupTestData(t, formatRandomResourceCollectionPath(testFolder), logger)
}

// Format a path to save RandomResourceCollection in the given folder
func formatRandomResourceCollectionPath(testFolder string) string {
	return FormatTestDataPath(testFolder, "RandomResourceCollection.json")
}

// Serialize and save an AMI ID into the given folder. This allows you to build an AMI during setup and to reuse that
// AMI later during validation and teardown.
func SaveAmiId(t *testing.T, testFolder string, amiId string, logger *log.Logger) {
	SaveTestData(t, formatAmiIdPath(testFolder), amiId, logger)
}

// Load and unserialize an AMI ID from the given folder. This allows you to reuse an AMI  that was created during an
// earlier setup step in later validation and teardown steps.
func LoadAmiId(t *testing.T, testFolder string, logger *log.Logger) string {
	var amiId string
	LoadTestData(t, formatAmiIdPath(testFolder), &amiId, logger)
	return amiId
}

// Clean up the files used to store an AMI ID between test stages
func CleanupAmiId(t *testing.T, testFolder string, logger *log.Logger) {
	CleanupTestData(t, formatAmiIdPath(testFolder), logger)
}

// Format a path to save an AMI ID in the given folder
func formatAmiIdPath(testFolder string) string {
	return FormatTestDataPath(testFolder, "AMI.json")
}

// Format a path to save test data
func FormatTestDataPath(testFolder string, filename string) string {
	return filepath.Join(testFolder, ".test-data", filename)
}

// Serialize and save a value used at test time to the given path. This allows you to create some sort of test data
// (e.g., TerratestOptions) during setup and to reuse this data later during validation and teardown.
func SaveTestData(t *testing.T, path string, value interface{}, logger *log.Logger) {
	logger.Printf("Storing test data in %s so it can be reused later", path)

	bytes, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Failed to convert value %s to JSON: %v", path, err)
	}

	t.Logf("Marshalled JSON: %s", string(bytes))

	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0777); err != nil {
		t.Fatalf("Failed to create folder %s: %v", parentDir, err)
	}

	if err := ioutil.WriteFile(path, bytes, 0644); err != nil {
		t.Fatalf("Failed to save value %s: %v", path, err)
	}
}

// Load and unserialize a value stored at the given path. The value should be a pointer to a struct into which the
// value will be deserialized. This allows you to reuse some sort of test data (e.g., TerratestOptions) from earlier
// setup steps in later validation and teardown steps.
func LoadTestData(t *testing.T, path string, value interface{}, logger *log.Logger) {
	logger.Printf("Loading test data from %s", path)

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load value from %s: %v", path, err)
	}

	if err := json.Unmarshal(bytes, value); err != nil {
		t.Fatalf("Failed to parse JSON for value %s: %v", path, err)
	}
}

// Return true if a file exists at $path and the test data there is non-empty.
func IsTestDataPresent(t *testing.T, path string, logger *log.Logger) bool {
	logger.Printf("Testing whether test data exists at %s", path)

	bytes, err := ioutil.ReadFile(path)
	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		logger.Printf("No test data was found at %s", path)
		return false
	} else if err != nil {
		t.Fatalf("Failed to load test data from %s due to unexpected error: %v", path, err)
	}

	if isEmptyJson(t, bytes) {
		logger.Printf("No test data was found at %s", path)
		return false
	}

	logger.Printf("Non-empty test data found at %s", path)
	return true
}

// Return true if the given bytes are empty, or in a valid JSON format that can reasonably be considered empty.
// The types used are based on the type possibilities listed at https://golang.org/src/encoding/json/decode.go?s=4062:4110#L51
func isEmptyJson(t *testing.T, bytes []byte) bool {
	var value interface{}

	if len(bytes) == 0 {
		return true
	}

	if err := json.Unmarshal(bytes, &value); err != nil {
		t.Fatalf("Failed to parse JSON while testing whether it is empty: %v", err)
	}

	if value == nil {
		return true
	}

	valueBool, ok := value.(bool)
	if ok && ! valueBool {
		return true
	}

	valueFloat64, ok := value.(float64)
	if ok && valueFloat64 == 0 {
		return true
	}

	valueString, ok := value.(string)
	if ok && valueString == "" {
		return true
	}

	valueSlice, ok := value.([]interface{})
	if ok && len(valueSlice) == 0 {
		return true
	}

	valueMap, ok := value.(map[string]interface{})
	if ok && len(valueMap) == 0 {
		return true
	}

	return false
}

// Clean up the test data at the given path
func CleanupTestData(t *testing.T, path string, logger *log.Logger) {
	if files.FileExists(path) {
		logger.Printf("Cleaning up test data from %s", path)
		if err := os.Remove(path); err != nil {
			t.Fatalf("Failed to clean up file at %s: %v", path, err)
		}
	} else {
		logger.Printf("%s does not exist. Nothing to cleanup.", path)
	}
}

// Copy the given root folder to a randomly-named temp folder and return the path to the given examples folder within
// the new temp root folder. This is useful when running multiple tests in parallel against the same set of Terraform
// files to ensure the tests don't overwrite each other's .terraform working directory and terraform.tfstate files. To
// ensure relative paths work, we copy over the entire root folder to a temp folder, and then return the path within
// that temp folder to the given example dir, which is where the actual test will be running.
//
// Note that if any of the SKIP_<stage> environment variables is set, we assume this is a test in the local dev where
// there are no other concurrent tests running and we want to be able to cache test data between test stages, so in
// that case, we do NOT copy anything to a temp folder, an dreturn the path to the original examples folder instead.
func CopyTerraformFolderToTemp(t *testing.T, rootFolder string, examplesFolder string, testName string, logger *log.Logger) string {
	if SkipStageEnvVarSet() {
		logger.Printf("A SKIP_XXX environment variable is set. Using original examples folder rather than a temp folder so we can cache data between stages for faster local testing.")
		return filepath.Join(rootFolder, examplesFolder)
	}

	tmpRootFolder, err := files.CopyTerraformFolderToTemp(rootFolder, testName)
	if err != nil {
		t.Fatal(err)
	}

	return filepath.Join(tmpRootFolder, examplesFolder)
}