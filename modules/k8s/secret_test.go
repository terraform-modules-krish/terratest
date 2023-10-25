// +build kubernetes

// NOTE: we have build tags to differentiate kubernetes tests from non-kubernetes tests. This is done because minikube
// is heavy and can interfere with docker related tests in terratest. To avoid overloading the system, we run the
// kubernetes tests separately from the others.

package k8s

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/terraform-modules-krish/terratest/modules/random"
)

func TestGetSecretEReturnsErrorForNonExistantSecret(t *testing.T) {
	t.Parallel()

	options := NewKubectlOptions("", "")
	_, err := GetSecretE(t, options, "master-password")
	require.Error(t, err)
}

func TestGetSecretEReturnsCorrectSecretInCorrectNamespace(t *testing.T) {
	t.Parallel()

	uniqueID := strings.ToLower(random.UniqueId())
	options := NewKubectlOptions("", "")
	options.Namespace = uniqueID
	configData := fmt.Sprintf(EXAMPLE_SECRET_YAML_TEMPLATE, uniqueID, uniqueID)
	defer KubectlDeleteFromString(t, options, configData)
	KubectlApplyFromString(t, options, configData)

	secret := GetSecret(t, options, "master-password")
	require.Equal(t, secret.Name, "master-password")
	require.Equal(t, secret.Namespace, uniqueID)
}

const EXAMPLE_SECRET_YAML_TEMPLATE = `---
apiVersion: v1
kind: Namespace
metadata:
  name: %s
---
apiVersion: v1
kind: Secret
metadata:
  name: master-password
  namespace: %s
`
