package gcp

import (
	"context"
	"fmt"
	"path"
	"testing"

	"github.com/terraform-modules-krish/terratest/modules/logger"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

// GetPublicIPOfInstance gets the public IP address of the given Instance in the given region.
func GetPublicIPOfInstance(t *testing.T, projectID string, zone string, instanceID string) string {
	ip, err := GetPublicIPOfInstanceE(t, projectID, zone, instanceID)
	if err != nil {
		t.Fatal(err)
	}
	return ip
}

// GetPublicIPOfInstanceE gets the public IP address of the given Instance in the given region.
func GetPublicIPOfInstanceE(t *testing.T, projectID string, zone string, instanceID string) (string, error) {
	logger.Logf(t, "Getting Public IP for Compute Instance %s", instanceID)

	ctx := context.Background()

	service, err := NewComputeServiceE(t)
	if err != nil {
		return "", err
	}

	instance, err := service.Instances.Get(projectID, zone, instanceID).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("Instances.Get(%s) got error: %v", instanceID, err)
	}

	// If there are no accessConfigs specified, then this instance will have no external internet access:
	// https://cloud.google.com/compute/docs/reference/rest/v1/instances.
	if len(instance.NetworkInterfaces[0].AccessConfigs) == 0 {
		return "", fmt.Errorf("Attempted to get public IP of Compute Instance %s, but that Compute Instance does not have a public IP address", instanceID)
	}

	ip := instance.NetworkInterfaces[0].AccessConfigs[0].NatIP

	return ip, nil
}

// GetLabelsForComputeInstance returns all the tags for the given Compute Instance.
func GetLabelsForComputeInstance(t *testing.T, projectID string, zone string, instanceID string) map[string]string {
	labels, err := GetLabelsForComputeInstanceE(t, projectID, zone, instanceID)
	if err != nil {
		t.Fatal(err)
	}
	return labels
}

// GetLabelsForComputeInstanceE returns all the tags for the given Compute Instance.
func GetLabelsForComputeInstanceE(t *testing.T, projectID string, zone string, instanceID string) (map[string]string, error) {
	logger.Logf(t, "Getting Labels for Compute Instance %s", instanceID)

	ctx := context.Background()

	service, err := NewComputeServiceE(t)
	if err != nil {
		return nil, err
	}

	instance, err := service.Instances.Get(projectID, zone, instanceID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("Instances.Get(%s) got error: %v", instanceID, err)
	}

	return instance.Labels, nil
}

// DeleteImage deletes the given Compute Image.
func DeleteImage(t *testing.T, projectID string, imageID string) {
	err := DeleteImageE(t, projectID, imageID)
	if err != nil {
		t.Fatal(err)
	}
}

// DeleteImageE deletes the given Compute Image.
func DeleteImageE(t *testing.T, projectID string, imageID string) error {
	logger.Logf(t, "Destroying Image %s", imageID)

	ctx := context.Background()

	service, err := NewComputeServiceE(t)
	if err != nil {
		return err
	}

	if _, err := service.Images.Delete(projectID, imageID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("Images.Delete(%s) got error: %v", imageID, err)
	}

	return err
}

// AddLabelsToInstance adds the tags to the given taggable instance.
func AddLabelsToInstance(t *testing.T, projectID string, zone string, instance string, labels map[string]string) {
	err := AddLabelsToInstanceE(t, projectID, zone, instance, labels)
	if err != nil {
		t.Fatal(err)
	}
}

// AddLabelsToInstanceE adds the tags to the given taggable instance.
func AddLabelsToInstanceE(t *testing.T, projectID string, zone string, instance string, labels map[string]string) error {
	logger.Logf(t, "Adding labels to instance %s in zone %s", instance, zone)

	ctx := context.Background()

	service, err := NewComputeServiceE(t)
	if err != nil {
		return err
	}

	// Get the fingerprint of the existing labels
	existingInstance, err := service.Instances.Get(projectID, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("Instances.Get(%s) got error: %v", instance, err)
	}
	req := compute.InstancesSetLabelsRequest{Labels: labels, LabelFingerprint: existingInstance.LabelFingerprint}

	// Perform the SetLabels request
	if _, err := service.Instances.SetLabels(projectID, zone, instance, &req).Context(ctx).Do(); err != nil {
		return fmt.Errorf("Instances.SetLabels(%s) got error: %v", instance, err)
	}

	return err
}

// GetInstanceIdsForInstanceGroup gets the IDs of Instances in the given Instance Group.
func GetInstanceIdsForInstanceGroup(t *testing.T, projectID string, zone string, groupName string) []string {
	ids, err := GetInstanceIdsForInstanceGroupE(t, projectID, zone, groupName)
	if err != nil {
		t.Fatal(err)
	}
	return ids
}

// GetInstanceIdsForInstanceGroupE gets the IDs of Instances in the given Instance Group.
func GetInstanceIdsForInstanceGroupE(t *testing.T, projectID string, zone string, groupName string) ([]string, error) {
	logger.Logf(t, "Get instances for instance group %s in zone %s", groupName, zone)

	ctx := context.Background()

	service, err := NewComputeServiceE(t)
	if err != nil {
		return nil, err
	}

	requestBody := &compute.InstanceGroupsListInstancesRequest{
		InstanceState: "ALL",
	}

	instanceIDs := []string{}
	req := service.InstanceGroups.ListInstances(projectID, zone, groupName, requestBody)
	if err := req.Pages(ctx, func(page *compute.InstanceGroupsListInstances) error {
		for _, instance := range page.Items {
			// For some reason service.InstanceGroups.ListInstances returns us a collection
			// with Instance URLs and we need only the Instance ID for the next call. Use
			// the path functions to chop the Instance ID off the end of the URL.
			instanceID := path.Base(instance.Instance)
			instanceIDs = append(instanceIDs, instanceID)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("InstanceGroups.ListInstances(%s) got error: %v", groupName, err)
	}

	return instanceIDs, nil
}

// NewComputeService creates a new Compute service.
func NewComputeService(t *testing.T) *compute.Service {
	client, err := NewComputeServiceE(t)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

// NewComputeServiceE creates a new Compute service.
func NewComputeServiceE(t *testing.T) (*compute.Service, error) {
	ctx := context.Background()

	client, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("Failed to get default client: %v", err)
	}

	service, err := compute.New(client)
	if err != nil {
		return nil, err
	}

	return service, nil
}
