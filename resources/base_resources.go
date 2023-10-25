package resources

import (
	"testing"
	"github.com/terraform-modules-krish/terratest"
)

func CreateBaseRandomResourceCollection(t *testing.T, excludedRegions ...string) *terratest.RandomResourceCollection {
	resourceCollectionOptions := terratest.NewRandomResourceCollectionOptions()

	if (excludedRegions!=nil) {
		resourceCollectionOptions.ForbiddenRegions = excludedRegions
	}

	randomResourceCollection, err := terratest.CreateRandomResourceCollection(resourceCollectionOptions)
	if err != nil {
		t.Fatalf("Failed to create random resource collection: %s\n", err.Error())
	}

	return randomResourceCollection
}
