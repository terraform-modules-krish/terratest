package gcp

import (
	"context"
	"os"
	"testing"

	"github.com/terraform-modules-krish/terratest/modules/collections"
	"github.com/terraform-modules-krish/terratest/modules/logger"
	"github.com/terraform-modules-krish/terratest/modules/random"
	compute "google.golang.org/api/compute/v1"
)

// You can set this environment variable to force Terratest to use a specific region rather than a random one. This is
// convenient when iterating locally.
const regionOverrideEnvVarName = "TERRATEST_GCP_REGION"

// You can set this environment variable to force Terratest to use a specific zone rather than a random one. This is
// convenient when iterating locally.
const zoneOverrideEnvVarName = "TERRATEST_GCP_ZONE"

// Some GCP API calls require a GCP region. We typically require the user to set one explicitly, but in some
// cases, this doesn't make sense (e.g., for fetching the list of regions in an account), so for those cases, we use
// this region as a default.
const defaultRegion = "us-west1"

// Some GCP API calls require a GCP zone. We typically require the user to set one explicitly, but in some
// cases, this doesn't make sense (e.g., for fetching the list of regions in an account), so for those cases, we use
// this zone as a default.
const defaultZone = "us-west1-b"

// GetRandomRegion gets a randomly chosen GCP region. If approvedRegions is not empty, this will be a region from the approvedRegions
// list; otherwise, this method will fetch the latest list of regions from the GCP APIs and pick one of those. If
// forbiddenRegions is not empty, this method will make sure the returned region is not in the forbiddenRegions list.
func GetRandomRegion(t *testing.T, approvedRegions []string, forbiddenRegions []string) string {
	region, err := GetRandomRegionE(t, approvedRegions, forbiddenRegions)
	if err != nil {
		t.Fatal(err)
	}
	return region
}

// GetRandomRegionE gets a randomly chosen GCP region. If approvedRegions is not empty, this will be a region from the approvedRegions
// list; otherwise, this method will fetch the latest list of regions from the GCP APIs and pick one of those. If
// forbiddenRegions is not empty, this method will make sure the returned region is not in the forbiddenRegions list.
func GetRandomRegionE(t *testing.T, approvedRegions []string, forbiddenRegions []string) (string, error) {
	regionFromEnvVar := os.Getenv(regionOverrideEnvVarName)
	if regionFromEnvVar != "" {
		logger.Logf(t, "Using GCP region %s from environment variable %s", regionFromEnvVar, regionOverrideEnvVarName)
		return regionFromEnvVar, nil
	}

	regionsToPickFrom := approvedRegions

	if len(regionsToPickFrom) == 0 {
		allRegions, err := GetAllGcpRegionsE(t)
		if err != nil {
			return "", err
		}
		regionsToPickFrom = allRegions
	}

	regionsToPickFrom = collections.ListSubtract(regionsToPickFrom, forbiddenRegions)
	region := random.RandomString(regionsToPickFrom)

	logger.Logf(t, "Using region %s", region)
	return region, nil
}

// GetRandomZone gets a randomly chosen GCP zone. If approvedRegions is not empty, this will be a zone from the approvedZones
// list; otherwise, this method will fetch the latest list of zones from the GCP APIs and pick one of those. If
// forbiddenZones is not empty, this method will make sure the returned region is not in the forbiddenZones list.
func GetRandomZone(t *testing.T, approvedZones []string, forbiddenZones []string) string {
	zone, err := GetRandomZoneE(t, approvedZones, forbiddenZones)
	if err != nil {
		t.Fatal(err)
	}
	return zone
}

// GetRandomZoneE gets a randomly chosen GCP zone. If approvedRegions is not empty, this will be a zone from the approvedZones
// list; otherwise, this method will fetch the latest list of zones from the GCP APIs and pick one of those. If
// forbiddenZones is not empty, this method will make sure the returned region is not in the forbiddenZones list.
func GetRandomZoneE(t *testing.T, approvedZones []string, forbiddenZones []string) (string, error) {
	zoneFromEnvVar := os.Getenv(zoneOverrideEnvVarName)
	if zoneFromEnvVar != "" {
		logger.Logf(t, "Using GCP zone %s from environment variable %s", zoneFromEnvVar, zoneOverrideEnvVarName)
		return zoneFromEnvVar, nil
	}

	zonesToPickFrom := approvedZones

	if len(zonesToPickFrom) == 0 {
		allZones, err := GetAllGcpZonesE(t)
		if err != nil {
			return "", err
		}
		zonesToPickFrom = allZones
	}

	zonesToPickFrom = collections.ListSubtract(zonesToPickFrom, forbiddenZones)
	zone := random.RandomString(zonesToPickFrom)

	logger.Logf(t, "Using zone %s", zone)
	return zone, nil
}

// GetAllGcpRegions gets the list of GCP regions available in this account.
func GetAllGcpRegions(t *testing.T) []string {
	out, err := GetAllGcpRegionsE(t)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// GetAllGcpRegionsE gets the list of GCP regions available in this account.
func GetAllGcpRegionsE(t *testing.T) ([]string, error) {
	logger.Log(t, "Looking up all GCP regions available in this account")

	// TODO - NewComputeServiceE creates a context, should we get that somehow
	// or use a new one here?
	ctx := context.Background()

	service, err := NewComputeServiceE(t)
	if err != nil {
		return nil, err
	}

	// Project ID for this request.
	// TODO - should we read the Project ID or pass it in?
	projectID := GetGoogleProjectIDFromEnvVar()

	req := service.Regions.List(projectID)

	regions := []string{}
	if err := req.Pages(ctx, func(page *compute.RegionList) error {
		for _, region := range page.Items {
			regions = append(regions, region.Name)
		}
		return err
	}); err != nil {
		return nil, err
	}

	return regions, nil
}

// GetAllGcpZones gets the list of GCP zones available in this account.
func GetAllGcpZones(t *testing.T) []string {
	out, err := GetAllGcpZonesE(t)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// GetAllGcpZonesE gets the list of GCP zones available in this account.
func GetAllGcpZonesE(t *testing.T) ([]string, error) {
	logger.Log(t, "Looking up all GCP zones available in this account")

	// TODO - NewComputeServiceE creates a context, should we get that somehow
	// or use a new one here?
	ctx := context.Background()

	service, err := NewComputeServiceE(t)
	if err != nil {
		return nil, err
	}

	// Project ID for this request.
	// TODO - should we read the Project ID or pass it in?
	projectID := GetGoogleProjectIDFromEnvVar()

	req := service.Zones.List(projectID)

	zones := []string{}
	if err := req.Pages(ctx, func(page *compute.ZoneList) error {
		for _, zone := range page.Items {
			zones = append(zones, zone.Name)
		}
		return err
	}); err != nil {
		return nil, err
	}

	return zones, nil
}
