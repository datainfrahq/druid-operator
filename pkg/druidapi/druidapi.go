package druidapi

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DruidRouterPort  = "8088"
	OperatorUserName = "OperatorUserName"
	OperatorPassword = "OperatorPassword"
)

// RouterConnectionInfo contains the inferred connection details for a Druid router
type RouterConnectionInfo struct {
	Protocol string
	Port     string
}

type AuthType string

const (
	BasicAuth AuthType = "basic-auth"
)

type Auth struct {
	// +required
	Type AuthType `json:"type"`
	// +required
	SecretRef v1.SecretReference `json:"secretRef"`

	// UsernameKey specifies the key within the Kubernetes secret that contains the username for authentication.
	UsernameKey string `json:"usernameKey,omitempty"`

	// PasswordKey specifies the key within the Kubernetes secret that contains the password for authentication.
	PasswordKey string `json:"passwordKey,omitempty"`
}

// GetAuthCreds retrieves basic authentication credentials from a Kubernetes secret.
// If the Auth object is empty, it returns an empty BasicAuth object.
// Parameters:
//
//	ctx: The context object.
//	c: The Kubernetes client.
//	auth: The Auth object containing the secret reference.
//
// Returns:
//
//	BasicAuth: The basic authentication credentials, or an error if authentication retrieval fails.
func GetAuthCreds(
	ctx context.Context,
	c client.Client,
	auth Auth,
) (internalhttp.BasicAuth, error) {
	userNameKey := OperatorUserName
	passwordKey := OperatorPassword

	if auth.UsernameKey != "" {
		userNameKey = auth.UsernameKey
	}

	if auth.PasswordKey != "" {
		passwordKey = auth.PasswordKey
	}

	// Check if the mentioned secret exists
	if auth != (Auth{}) {
		secret := v1.Secret{}
		if err := c.Get(ctx, types.NamespacedName{
			Namespace: auth.SecretRef.Namespace,
			Name:      auth.SecretRef.Name,
		}, &secret); err != nil {
			return internalhttp.BasicAuth{}, err
		}

		if _, ok := secret.Data[userNameKey]; !ok {
			return internalhttp.BasicAuth{}, fmt.Errorf("username key %q not found in secret %s/%s", userNameKey, auth.SecretRef.Namespace, auth.SecretRef.Name)
		}

		if _, ok := secret.Data[passwordKey]; !ok {
			return internalhttp.BasicAuth{}, fmt.Errorf("password key %q not found in secret %s/%s", passwordKey, auth.SecretRef.Namespace, auth.SecretRef.Name)
		}

		creds := internalhttp.BasicAuth{
			UserName: string(secret.Data[userNameKey]),
			Password: string(secret.Data[passwordKey]),
		}

		return creds, nil
	}

	return internalhttp.BasicAuth{}, nil
}

// parseProperties parses Java properties format string into a map
func parseProperties(props string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(props, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first '=' character
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
	}

	return result
}

// InferRouterConnectionFromConfig analyzes Druid configuration to determine router connection settings
func InferRouterConnectionFromConfig(commonConfig, routerConfig string) RouterConnectionInfo {
	// Parse both configurations
	commonProps := parseProperties(commonConfig)
	routerProps := parseProperties(routerConfig)

	// Merge configurations, with router config taking precedence
	mergedProps := make(map[string]string)
	for k, v := range commonProps {
		mergedProps[k] = v
	}
	for k, v := range routerProps {
		mergedProps[k] = v
	}

	// Default values
	protocol := "http"
	port := DruidRouterPort

	// Check TLS configuration
	enableTlsPort := mergedProps["druid.enableTlsPort"]
	enablePlaintextPort := mergedProps["druid.enablePlaintextPort"]
	tlsPort := mergedProps["druid.tlsPort"]
	plaintextPort := mergedProps["druid.plaintextPort"]

	// Determine protocol and port based on configuration
	if enableTlsPort == "true" {
		protocol = "https"
		if tlsPort != "" {
			port = tlsPort
		} else {
			// If TLS is enabled but no TLS port specified, check for druid.port
			if druidPort := mergedProps["druid.port"]; druidPort != "" {
				port = druidPort
			}
		}
	} else if enablePlaintextPort != "false" { // Default is true if not explicitly disabled
		protocol = "http"
		if plaintextPort != "" {
			port = plaintextPort
		} else {
			// Check for druid.port as fallback
			if druidPort := mergedProps["druid.port"]; druidPort != "" {
				port = druidPort
			}
		}
	}

	return RouterConnectionInfo{
		Protocol: protocol,
		Port:     port,
	}
}

// GetRouterConfigFromCluster retrieves router configuration from Kubernetes ConfigMaps
func GetRouterConfigFromCluster(ctx context.Context, c client.Client, namespace, druidClusterName string) (RouterConnectionInfo, error) {
	// Get common configuration
	commonConfigMapName := fmt.Sprintf("%s-druid-common-config", druidClusterName)
	commonConfigMap := &v1.ConfigMap{}
	err := c.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      commonConfigMapName,
	}, commonConfigMap)
	if err != nil {
		return RouterConnectionInfo{}, fmt.Errorf("failed to get common config map %s: %w", commonConfigMapName, err)
	}

	commonConfig := ""
	if commonRuntimeProps, exists := commonConfigMap.Data["common.runtime.properties"]; exists {
		commonConfig = commonRuntimeProps
	}

	// Get router-specific configuration
	routerConfigMapName := fmt.Sprintf("druid-%s-routers-config", druidClusterName)
	routerConfigMap := &v1.ConfigMap{}
	err = c.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      routerConfigMapName,
	}, routerConfigMap)

	routerConfig := ""
	if err == nil {
		// Router config map exists, get runtime properties
		if routerRuntimeProps, exists := routerConfigMap.Data["runtime.properties"]; exists {
			routerConfig = routerRuntimeProps
		}
	}
	// If router config map doesn't exist, that's fine - we'll just use common config

	return InferRouterConnectionFromConfig(commonConfig, routerConfig), nil
}

// MakePath constructs the appropriate path for the specified Druid API.
// Parameters:
//
//	baseURL: The base URL of the Druid cluster. For example, http://router-svc.namespace.svc.cluster.local:8088.
//	componentType: The type of Druid component. For example, "indexer".
//	apiType: The type of Druid API. For example, "worker".
//	additionalPaths: Additional path components to be appended to the URL.
//
// Returns:
//
//	string: The constructed path.
func MakePath(baseURL, componentType, apiType string, additionalPaths ...string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return ""
	}

	// Construct the initial path
	u.Path = path.Join("druid", componentType, "v1", apiType)

	// Append additional path components
	for _, p := range additionalPaths {
		u.Path = path.Join(u.Path, p)
	}

	return u.String()
}

// GetRouterSvcUrl retrieves the URL of the Druid router service using configuration inference.
// Parameters:
//
//	ctx: The context object.
//	namespace: The namespace of the Druid cluster.
//	druidClusterName: The name of the Druid cluster.
//	c: The Kubernetes client.
//
// Returns:
//
//	string: The URL of the Druid router service.
func GetRouterSvcUrl(ctx context.Context, namespace, druidClusterName string, c client.Client) (string, error) {
	// Get router configuration from cluster ConfigMaps
	routerInfo, err := GetRouterConfigFromCluster(ctx, c, namespace, druidClusterName)
	if err != nil {
		return "", fmt.Errorf("failed to infer router configuration: %w", err)
	}

	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{
			"druid_cr":  druidClusterName,
			"component": "router",
		}),
	}
	svcList := &v1.ServiceList{}
	if err := c.List(ctx, svcList, listOpts...); err != nil {
		return "", err
	}
	var svcName string

	for range svcList.Items {
		svcName = svcList.Items[0].Name
	}

	if svcName == "" {
		return "", errors.New("router svc discovery fail")
	}

	newName := routerInfo.Protocol + "://" + svcName + "." + namespace + ":" + routerInfo.Port

	return newName, nil
}
