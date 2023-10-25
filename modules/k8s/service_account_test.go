package k8s

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/terraform-modules-krish/terratest/modules/random"
)

func TestGetServiceAccountEReturnsErrorForNonExistantServiceAccount(t *testing.T) {
	t.Parallel()

	options := NewKubectlOptions("", "")
	_, err := GetServiceAccountE(t, options, "terratest")
	require.Error(t, err)
}

func TestGetServiceAccountEReturnsCorrectServiceAccountInCorrectNamespace(t *testing.T) {
	t.Parallel()

	uniqueID := strings.ToLower(random.UniqueId())
	options := NewKubectlOptions("", "")
	options.Namespace = uniqueID
	configData := fmt.Sprintf(EXAMPLE_SERVICEACCOUNT_YAML_TEMPLATE, uniqueID, uniqueID)
	defer KubectlDeleteFromString(t, options, configData)
	KubectlApplyFromString(t, options, configData)

	serviceAccount := GetServiceAccount(t, options, "terratest")
	require.Equal(t, serviceAccount.Name, "terratest")
	require.Equal(t, serviceAccount.Namespace, uniqueID)
}

func TestCreateServiceAccountECreatesServiceAccountInNamespaceWithGivenName(t *testing.T) {
	t.Parallel()

	uniqueID := strings.ToLower(random.UniqueId())
	options := NewKubectlOptions("", "")
	options.Namespace = uniqueID
	defer DeleteNamespace(t, options, options.Namespace)
	CreateNamespace(t, options, options.Namespace)

	// Note: We don't need to delete this at the end of test, because deleting the namespace automatically deletes
	// everything created in the namespace.
	CreateServiceAccount(t, options, "terratest")
	serviceAccount := GetServiceAccount(t, options, "terratest")
	require.Equal(t, serviceAccount.Name, "terratest")
	require.Equal(t, serviceAccount.Namespace, uniqueID)
}

const EXAMPLE_SERVICEACCOUNT_YAML_TEMPLATE = `---
apiVersion: v1
kind: Namespace
metadata:
  name: %s
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: terratest
  namespace: %s
`
