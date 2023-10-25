package k8s

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	http_helper "github.com/terraform-modules-krish/terratest/modules/http-helper"
	"github.com/terraform-modules-krish/terratest/modules/random"
)

func TestGetServiceEReturnsErrorForNonExistantService(t *testing.T) {
	t.Parallel()

	options := NewKubectlOptions("", "")
	_, err := GetServiceE(t, options, "nginx-service")
	require.Error(t, err)
}

func TestGetServiceEReturnsCorrectServiceInCorrectNamespace(t *testing.T) {
	t.Parallel()

	uniqueID := strings.ToLower(random.UniqueId())
	options := NewKubectlOptions("", "")
	options.Namespace = uniqueID
	configData := fmt.Sprintf(EXAMPLE_DEPLOYMENT_YAML_TEMPLATE, uniqueID, uniqueID, uniqueID)
	KubectlApplyFromString(t, options, configData)
	defer KubectlDeleteFromString(t, options, configData)

	service := GetService(t, options, "nginx-service")
	require.Equal(t, service.Name, "nginx-service")
	require.Equal(t, service.Namespace, uniqueID)
}

func TestWaitUntilServiceAvailableReturnsSuccessfullyOnNodePortType(t *testing.T) {
	t.Parallel()

	uniqueID := strings.ToLower(random.UniqueId())
	options := NewKubectlOptions("", "")
	options.Namespace = uniqueID
	configData := fmt.Sprintf(EXAMPLE_DEPLOYMENT_YAML_TEMPLATE, uniqueID, uniqueID, uniqueID)
	KubectlApplyFromString(t, options, configData)
	defer KubectlDeleteFromString(t, options, configData)

	WaitUntilServiceAvailable(t, options, "nginx-service", 10, 1*time.Second)
}

func TestGetServiceEndpointEReturnsAccessibleEndpointForNodePort(t *testing.T) {
	t.Parallel()

	uniqueID := strings.ToLower(random.UniqueId())
	options := NewKubectlOptions("", "")
	options.Namespace = uniqueID
	configData := fmt.Sprintf(EXAMPLE_DEPLOYMENT_YAML_TEMPLATE, uniqueID, uniqueID, uniqueID)
	KubectlApplyFromString(t, options, configData)
	defer KubectlDeleteFromString(t, options, configData)

	service := GetService(t, options, "nginx-service")
	endpoint := GetServiceEndpoint(t, service, 80)
	// Test up to 5 minutes
	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		fmt.Sprintf("http://%s", endpoint),
		30,
		10*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)
}

const EXAMPLE_DEPLOYMENT_YAML_TEMPLATE = `---
apiVersion: v1
kind: Namespace
metadata:
  name: %s
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  namespace: %s
spec:
  selector:
    matchLabels:
      app: nginx
  replicas: 1
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.15.7
        ports:
        - containerPort: 80
---
kind: Service
apiVersion: v1
metadata:
  name: nginx-service
  namespace: %s
spec:
  selector:
    app: nginx
  ports:
  - protocol: TCP
    targetPort: 80
    port: 80
  type: NodePort
`
