package terratest

import (
	"testing"
	"path"
	"strconv"
	"errors"
	"time"
	"github.com/terraform-modules-krish/terratest/test-util"
)

const DOMAIN_KEY = "domain"
const PORT_KEY = "port"
const TEST_SERVER_DOMAIN = "0.0.0.0"
const TEST_SERVER_TEXT = "Hello, World"

func TestUrlCheckerWithDummyServer(t *testing.T) {
	t.Parallel()

	randomResourceCollectionOptions := NewRandomResourceCollectionOptions()
	randomResourceCollection, err := CreateRandomResourceCollection(randomResourceCollectionOptions)
	defer randomResourceCollection.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	listener, port, err := test_util.RunDummyServer(TEST_SERVER_TEXT)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	options := NewTerratestOptions()
	options.UniqueId = randomResourceCollection.UniqueId
	options.TestName = "Test - TestUrlCheckerWithDummyServer"
	options.TemplatePath = path.Join(fixtureDir, "url-checker-with-server-passthrough")
	options.Vars = map[string]string{DOMAIN_KEY: TEST_SERVER_DOMAIN, PORT_KEY: strconv.Itoa(port)}

	defer Destroy(options, randomResourceCollection)
	if _, err := Apply(options); err != nil {
		t.Fatal(err)
	}

	if err := CheckTerraformOutputUrlReturnsExpectedTextWithinTimeLimit(options, DOMAIN_KEY, PORT_KEY, TEST_SERVER_TEXT, 5, 1 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestUrlCheckerWithoutDummyServer(t *testing.T) {
	t.Parallel()

	randomResourceCollectionOptions := NewRandomResourceCollectionOptions()
	randomResourceCollection, err := CreateRandomResourceCollection(randomResourceCollectionOptions)
	defer randomResourceCollection.DestroyResources()
	if err != nil {
		t.Errorf("Failed to create random resource collection: %s\n", err.Error())
	}

	options := NewTerratestOptions()
	options.UniqueId = randomResourceCollection.UniqueId
	options.TestName = "Test - TestUrlCheckerWithoutDummyServer"
	options.TemplatePath = path.Join(fixtureDir, "url-checker-without-server-passthrough")
	options.Vars = map[string]string{DOMAIN_KEY: TEST_SERVER_DOMAIN, PORT_KEY: "12345"}

	defer Destroy(options, randomResourceCollection)
	if _, err := Apply(options); err != nil {
		t.Fatal(err)
	}

	if err := CheckTerraformOutputUrlReturnsExpectedTextWithinTimeLimit(options, DOMAIN_KEY, PORT_KEY, TEST_SERVER_TEXT, 5, 1 * time.Second); err == nil {
		t.Fatal(errors.New("Expected to get an error when testing a URL that doesn't work, but got nil"))
	}
}