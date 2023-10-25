package k8s

import (
	"errors"
	"path/filepath"
	"sort"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/terraform-modules-krish/terratest/modules/environment"
	"github.com/terraform-modules-krish/terratest/modules/logger"
)

// LoadConfigFromPath will load a ClientConfig object from a file path that points to a location on disk containing a
// kubectl config.
func LoadConfigFromPath(path string) clientcmd.ClientConfig {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: path},
		&clientcmd.ConfigOverrides{})
	return config
}

// DeleteConfigContextE will remove the context specified at the provided name, and remove any clusters and authinfos
// that are orphaned as a result of it. The config path is either specified in the environment variable KUBECONFIG or at
// the user's home directory under `.kube/config`.
func DeleteConfigContextE(t *testing.T, contextName string) error {
	kubeConfigPath, err := GetKubeConfigPathE(t)
	if err != nil {
		return err
	}

	return DeleteConfigContextWithPathE(t, kubeConfigPath, contextName)
}

// DeleteConfigContextWithPathE will remove the context specified at the provided name, and remove any clusters and
// authinfos that are orphaned as a result of it.
func DeleteConfigContextWithPathE(t *testing.T, kubeConfigPath string, contextName string) error {
	logger.Logf(t, "Removing kubectl config context %s from config at path %s", contextName, kubeConfigPath)

	// Load config and get data structure representing config info
	config := LoadConfigFromPath(kubeConfigPath)
	rawConfig, err := config.RawConfig()
	if err != nil {
		return err
	}

	// Check if the context we want to delete actually exists, and if so, delete it.
	_, ok := rawConfig.Contexts[contextName]
	if !ok {
		logger.Logf(t, "WARNING: Could not find context %s from config at path %s", contextName, kubeConfigPath)
		return nil
	}
	delete(rawConfig.Contexts, contextName)

	// If the removing context is the current context, be sure to set a new one
	if contextName == rawConfig.CurrentContext {
		if err := setNewContext(&rawConfig); err != nil {
			return err
		}
	}

	// Finally, clean up orphaned clusters and authinfos and then save config
	RemoveOrphanedClusterAndAuthInfoConfig(&rawConfig)
	if err := clientcmd.ModifyConfig(config.ConfigAccess(), rawConfig, false); err != nil {
		return err
	}

	logger.Logf(
		t,
		"Removed context %s from config at path %s and any orphaned clusters and authinfos",
		contextName,
		kubeConfigPath)
	return nil
}

// setNewContext will pick the alphebetically first available context from the list of contexts in the config to use as
// the new current context
func setNewContext(config *api.Config) error {
	// Sort contextNames and pick the first one
	var contextNames []string
	for name := range config.Contexts {
		contextNames = append(contextNames, name)
	}
	sort.Strings(contextNames)
	if len(contextNames) > 0 {
		config.CurrentContext = contextNames[0]
	} else {
		return errors.New("There are no available contexts remaining")
	}
	return nil
}

// RemoveOrphanedClusterAndAuthInfoConfig will remove all configurations related to clusters and users that have no
// contexts associated with it
func RemoveOrphanedClusterAndAuthInfoConfig(config *api.Config) {
	newAuthInfos := map[string]*api.AuthInfo{}
	newClusters := map[string]*api.Cluster{}
	for _, context := range config.Contexts {
		newClusters[context.Cluster] = config.Clusters[context.Cluster]
		newAuthInfos[context.AuthInfo] = config.AuthInfos[context.AuthInfo]
	}
	config.AuthInfos = newAuthInfos
	config.Clusters = newClusters
}

// GetKubeConfigPathE determines which file path to use as the kubectl config path
func GetKubeConfigPathE(t *testing.T) (string, error) {
	kubeConfigPath := environment.GetFirstNonEmptyEnvVarOrEmptyString(t, []string{"KUBECONFIG"})
	if kubeConfigPath == "" {
		configPath, err := KubeConfigPathFromHomeDirE()
		if err != nil {
			return "", err
		}
		kubeConfigPath = configPath
	}
	return kubeConfigPath, nil
}

// KubeConfigPathFromHomeDirE returns a string to the default Kubernetes config path in the home directory. This will
// error if the home directory can not be determined.
func KubeConfigPathFromHomeDirE() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(home, ".kube", "config")
	return configPath, err
}
