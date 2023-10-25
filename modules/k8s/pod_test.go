// +build kubernetes

// NOTE: we have build tags to differentiate kubernetes tests from non-kubernetes tests. This is done because minikube
// is heavy and can interfere with docker related tests in terratest. To avoid overloading the system, we run the
// kubernetes tests separately from the others.

package k8s

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/terraform-modules-krish/terratest/modules/random"
)

func TestGetPodEReturnsErrorForNonExistantPod(t *testing.T) {
	t.Parallel()

	options := NewKubectlOptions("", "")
	_, err := GetPodE(t, options, "nginx-pod")
	require.Error(t, err)
}

func TestGetPodEReturnsCorrectPodInCorrectNamespace(t *testing.T) {
	t.Parallel()

	uniqueID := strings.ToLower(random.UniqueId())
	options := NewKubectlOptions("", "")
	options.Namespace = uniqueID
	configData := fmt.Sprintf(EXAMPLE_POD_YAML_TEMPLATE, uniqueID, uniqueID)
	defer KubectlDeleteFromString(t, options, configData)
	KubectlApplyFromString(t, options, configData)

	pod := GetPod(t, options, "nginx-pod")
	require.Equal(t, pod.Name, "nginx-pod")
	require.Equal(t, pod.Namespace, uniqueID)
}

func TestWaitUntilPodAvailableReturnsSuccessfully(t *testing.T) {
	t.Parallel()

	uniqueID := strings.ToLower(random.UniqueId())
	options := NewKubectlOptions("", "")
	options.Namespace = uniqueID
	configData := fmt.Sprintf(EXAMPLE_POD_YAML_TEMPLATE, uniqueID, uniqueID)
	defer KubectlDeleteFromString(t, options, configData)
	KubectlApplyFromString(t, options, configData)

	WaitUntilPodAvailable(t, options, "nginx-pod", 60, 1*time.Second)
}

const EXAMPLE_POD_YAML_TEMPLATE = `---
apiVersion: v1
kind: Namespace
metadata:
  name: %s
---
apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
  namespace: %s
spec:
  containers:
  - name: nginx
    image: nginx:1.15.7
    ports:
    - containerPort: 80
`
