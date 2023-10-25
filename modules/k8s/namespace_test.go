package k8s

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/terraform-modules-krish/terratest/modules/random"
)

func TestNamespaces(t *testing.T) {
	t.Parallel()

	uniqueId := random.UniqueId()
	namespaceName := strings.ToLower(uniqueId)
	options := NewKubectlOptions("", "")
	CreateNamespace(t, options, namespaceName)
	defer func() {
		DeleteNamespace(t, options, namespaceName)
		namespace := GetNamespace(t, options, namespaceName)
		require.Equal(t, namespace.Status.Phase, corev1.NamespaceTerminating)
	}()

	namespace := GetNamespace(t, options, namespaceName)
	require.Equal(t, namespace.Name, namespaceName)
}
