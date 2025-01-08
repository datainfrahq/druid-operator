package druidapi

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"

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

type AuthType string

const (
	BasicAuth AuthType = "basic-auth"
)

type Auth struct {
	// +required
	Type AuthType `json:"type"`
	// +required
	SecretRef v1.SecretReference `json:"secretRef"`
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
//	BasicAuth: The basic authentication credentials.
func GetAuthCreds(
	ctx context.Context,
	c client.Client,
	auth Auth,
) (internalhttp.BasicAuth, error) {
	// Check if the mentioned secret exists
	if auth != (Auth{}) {
		secret := v1.Secret{}
		if err := c.Get(ctx, types.NamespacedName{
			Namespace: auth.SecretRef.Namespace,
			Name:      auth.SecretRef.Name,
		}, &secret); err != nil {
			return internalhttp.BasicAuth{}, err
		}
		creds := internalhttp.BasicAuth{
			UserName: string(secret.Data[OperatorUserName]),
			Password: string(secret.Data[OperatorPassword]),
		}

		return creds, nil
	}

	return internalhttp.BasicAuth{}, nil
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

// GetRouterSvcUrl retrieves the URL of the Druid router service.
// Parameters:
//
//	namespace: The namespace of the Druid cluster.
//	druidClusterName: The name of the Druid cluster.
//	c: The Kubernetes client.
//
// Returns:
//
//	string: The URL of the Druid router service.
func GetRouterSvcUrl(namespace, druidClusterName string, c client.Client) (string, error) {
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{
			"druid_cr":  druidClusterName,
			"component": "router",
		}),
	}
	svcList := &v1.ServiceList{}
	if err := c.List(context.Background(), svcList, listOpts...); err != nil {
		return "", err
	}
	var svcName string

	for range svcList.Items {
		svcName = svcList.Items[0].Name
	}

	if svcName == "" {
		return "", errors.New("router svc discovery fail")
	}

	newName := "http://" + svcName + "." + namespace + ".svc.cluster.local:" + DruidRouterPort

	return newName, nil
}
