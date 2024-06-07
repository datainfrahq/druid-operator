package lookup

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"strings"
)

func getOverrideUrls() (map[types.NamespacedName]string, error) {
	urls := make(map[types.NamespacedName]string)

	// TODO set env var name as const
	overrideVars, ok := os.LookupEnv("ROUTER_OVERRIDE_VARS")
	if !ok {
		return urls, nil
	}

	for _, overrideVar := range strings.Split(overrideVars, "|") {
		key, url, err := parseOverrideUrl(overrideVar)
		if err != nil {
			return nil, err
		}

		_, replaced := replace(urls, key, url)

		if replaced {
			return nil, fmt.Errorf("duplicate url override for cluster %v/%v specified", key.Namespace, key.Name)
		}
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
