package lookup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sClient struct {
	client client.Client
}

func NewK8sClient(client client.Client) K8sClient {
	return K8sClient{
		client: client,
	}
}

func (c *K8sClient) FindLookups(ctx context.Context, reports map[types.NamespacedName]Report) (map[ClusterKey]map[LookupKey]Spec, error) {
	specs := &v1alpha1.DruidLookupList{}
	if err := c.client.List(ctx, specs); err != nil {
		return nil, err
	}

	lookupSpecsPerCluster := make(map[ClusterKey]map[LookupKey]Spec)
	for _, spec := range specs.Items {
		clusterKey := ClusterKey{
			Namespace: spec.Namespace,
			Cluster:   spec.Spec.DruidClusterName,
		}
		lookupKey := LookupKey{
			Tier: spec.Spec.Tier,
			Id:   spec.Spec.Id,
		}
		var lookupSpec interface{}
		if err := json.Unmarshal([]byte(spec.Spec.Spec), &lookupSpec); err != nil {
			setIfNotPresent(
				reports,
				spec.GetNamespacedName(),
				Report(NewErrorReport(
					fmt.Errorf(
						"lookup resource %v in cluster %v/%v contains invalid spec, should be JSON",
						spec.Name,
						clusterKey.Namespace,
						clusterKey.Cluster,
					),
				)),
			)
			continue
		}

		if lookupKey.Tier == "" {
			lookupKey.Tier = "__default"
		}

		if lookupSpecsPerCluster[clusterKey] == nil {
			lookupSpecsPerCluster[clusterKey] = make(map[LookupKey]Spec)
		}
		ls := lookupSpecsPerCluster[clusterKey]

		if _, replaced := replace(ls, lookupKey, Spec{
			name: spec.GetNamespacedName(),
			spec: lookupSpec,
		}); replaced {
			setIfNotPresent(
				reports,
				spec.GetNamespacedName(),
				Report(NewErrorReport(
					fmt.Errorf(
						"resource %v specifies duplicate lookup %v/%v in cluster %v/%v",
						spec.Name,
						lookupKey.Tier,
						lookupKey.Id,
						clusterKey.Namespace,
						clusterKey.Cluster,
					),
				)),
			)
			continue
		}
	}

	return lookupSpecsPerCluster, nil
}

func (c *K8sClient) FindDruidCluster(ctx context.Context) (map[ClusterKey]*DruidClient, []error, error) {
	httpClient := internalhttp.NewHTTPClient(&http.Client{}, &internalhttp.Auth{BasicAuth: internalhttp.BasicAuth{}})
	clusters := make(map[ClusterKey]*DruidClient)
	nonFatalErrors := make([]error, 0)

	overrides, err := getOverrideUrls()
	if err != nil {
		return nil, nil, err
	}

	routerServices := &v1.ServiceList{}
	listOpt := client.MatchingLabels(map[string]string{
		"app":       "druid",
		"component": "router",
	})
	if err := c.client.List(ctx, routerServices, listOpt); err != nil {
		return nil, nil, err
	}

	for _, service := range routerServices.Items {
		key := ClusterKey{
			Namespace: service.Namespace,
			Cluster:   service.Labels["druid_cr"],
		}

		port, found := findFirst(service.Spec.Ports, func(p v1.ServicePort) bool {
			return p.Name == "service-port"
		})
		if !found {
			nonFatalErrors = append(nonFatalErrors, fmt.Errorf(`could not find "service-port" of router service %v/%v`, key.Namespace, service.Name))
			continue
		}

		url := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", service.Name, service.Namespace, port.Port)
		if override, found := overrides[key]; found {
			url = override
		}

		cluster, err := NewCluster(url, httpClient)
		if err != nil {
			nonFatalErrors = append(nonFatalErrors, errors.Join(fmt.Errorf("could not create druid cluster client for cluster at %v", url), err))
			continue
		}

		_, replaced := replace(clusters, key, cluster)
		if replaced {
			nonFatalErrors = append(nonFatalErrors, fmt.Errorf("duplicate router services found for cluster %v/%v", key.Namespace, key.Cluster))
			continue
		}
	}

	return clusters, nonFatalErrors, nil
}

func (c *K8sClient) UpdateStatus(ctx context.Context, name types.NamespacedName, report Report) error {
	lookup := &v1alpha1.DruidLookup{}
	err := c.client.Get(ctx, name, lookup)
	if err != nil {
		return err
	}

	err = report.MergeStatus(&lookup.Status)
	if err != nil {
		return err
	}

	return c.client.Status().Update(ctx, lookup)
}
