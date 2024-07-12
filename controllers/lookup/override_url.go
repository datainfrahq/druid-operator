package lookup

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"strings"
)

const (
	RouterOverrideVarsEnvVar = "ROUTER_OVERRIDE_VARS"
)

func getOverrideUrls() (map[types.NamespacedName]string, error) {
	urls := make(map[types.NamespacedName]string)

	overrideVars, ok := os.LookupEnv(RouterOverrideVarsEnvVar)
	if !ok {
		return urls, nil
	}

	for _, overrideVar := range strings.Split(overrideVars, "|") {
		key, url, err := parseOverrideUrl(overrideVar)
		if err != nil {
			return nil, err
		}

		if _, ok := urls[key]; ok {
			return nil, fmt.Errorf("duplicate url override for cluster %v/%v specified", key.Namespace, key.Name)
		}

		urls[key] = url
	}

	return urls, nil
}

func parseOverrideUrl(value string) (types.NamespacedName, string, error) {
	namespaceCluster, url, found := strings.Cut(value, "=")
	if !found {
		return types.NamespacedName{}, "", fmt.Errorf("invalid url override, no '=' found: %s", value)
	}

	namespace, cluster, found := strings.Cut(namespaceCluster, "/")
	if !found {
		return types.NamespacedName{}, "", fmt.Errorf("invalid namespace cluster spec, no '/' found: %s", namespaceCluster)
	}

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      cluster,
	}

	return key, url, nil
}
