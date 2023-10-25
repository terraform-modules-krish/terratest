package test

import (
	"testing"

	"github.com/terraform-modules-krish/terratest/modules/gcp"
	"github.com/terraform-modules-krish/terratest/modules/packer"
)

// An example of how to test the Packer template in examples/packer-basic-example using Terratest.
func TestPackerGCPBasicExample(t *testing.T) {
	t.Parallel()

	// Get the Project Id to use
	projectID := gcp.GetGoogleProjectIDFromEnvVar()

	// Pick a random GCP zone to test in. This helps ensure your code works in all regions.
	zone := gcp.GetRandomZone(t, nil, nil)

	packerOptions := &packer.Options{
		// The path to where the Packer template is located
		Template: "../examples/packer-basic-example/build.json",

		// Variables to pass to our Packer build using -var options
		Vars: map[string]string{
			"gcp_project_id": projectID,
			"gcp_zone":       zone,
		},

		// Only build the Google Compute Image
		Only: "googlecompute",
	}

	// Make sure the Packer build completes successfully
	imageID := packer.BuildArtifact(t, packerOptions)

	// Delete the Image after we're done
	defer gcp.DeleteImage(t, projectID, imageID)
}
