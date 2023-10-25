package k8s

import "github.com/terraform-modules-krish/terratest/modules/testing"

// IsMinikubeE returns true if the underlying kubernetes cluster is Minikube. This is determined by getting the
// associated nodes and checking if:
// - there is only one node
// - the node is named "minikube"
func IsMinikubeE(t testing.TestingT, options *KubectlOptions) (bool, error) {
	nodes, err := GetNodesE(t, options)
	if err != nil {
		return false, err
	}
	return len(nodes) == 1 && nodes[0].GetName() == "minikube", nil
}
