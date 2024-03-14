package lookup

import (
	"fmt"
	"os"
	"strings"
)

func getOverrideUrls() (map[ClusterKey]string, error) {
	urls := make(map[ClusterKey]string)

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
			return nil, fmt.Errorf("duplicate url override for cluster %v/%v specified", key.Namespace, key.Cluster)
		}
	}

	return urls, nil
}

func parseOverrideUrl(value string) (ClusterKey, string, error) {
	namespaceCluster, url, found := strings.Cut(value, "=")
	if !found {
		return ClusterKey{}, "", fmt.Errorf("invalid url override, no '=' found: %s", value)
	}

	namespace, cluster, found := strings.Cut(namespaceCluster, "/")
	if !found {
		return ClusterKey{}, "", fmt.Errorf("invalid namespace cluster spec, no '/' found: %s", namespaceCluster)
	}

	key := ClusterKey{
		Namespace: namespace,
		Cluster:   cluster,
	}

	return key, url, nil
}
