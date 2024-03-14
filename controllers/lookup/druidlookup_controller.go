/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lookup

import (
	"context"
	"fmt"
	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	druidv1alpha1 "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DruidLookupReconciler reconciles a DruidLookup object
type DruidLookupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=druid.apache.org,resources=druidlookups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=druid.apache.org,resources=druidlookups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=druid.apache.org,resources=druidlookups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *DruidLookupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO add event recording logic

	desiredLookupsPerCluster, err := findDesiredLookups(ctx, r)
	if err != nil {
		return ctrl.Result{}, err
	}

	overrideUrls, err := getOverrideUrls()
	if err != nil {
		return ctrl.Result{}, err
	}

	clusterUrls, err := findDruidClusterUrls(ctx, r, overrideUrls)
	if err != nil {
		return ctrl.Result{}, err
	}

	clusters, err := buildClusters(clusterUrls)
	if err != nil {
		return ctrl.Result{}, err
	}

	for key, cl := range clusters {
		desiredLookups := desiredLookupsPerCluster[key]

		if err := cl.Reconcile(desiredLookups); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func findDesiredLookups(ctx context.Context, r client.Reader) (map[ClusterKey]map[LookupKey]string, error) {
	dls := &v1alpha1.DruidLookupList{}
	if err := r.List(ctx, dls); err != nil {
		return nil, err
	}

	lsPerCluster := make(map[ClusterKey]map[LookupKey]string)
	for _, dl := range dls.Items {
		clusterKey := ClusterKey{
			Namespace: dl.Namespace,
			Cluster:   dl.Spec.DruidClusterName,
		}
		lookupKey := LookupKey{
			Tier: dl.Spec.Tier,
			Id:   dl.Spec.Id,
		}
		spec := dl.Spec.Spec

		if lookupKey.Tier == "" {
			lookupKey.Tier = "__default"
		}

		if lsPerCluster[clusterKey] == nil {
			lsPerCluster[clusterKey] = make(map[LookupKey]string)
		}
		ls := lsPerCluster[clusterKey]

		if _, replaced := replace(ls, lookupKey, spec); replaced {
			return nil, fmt.Errorf("resource %v specifies duplicate lookup %v/%v in cluster %v/%v", dl.Name, lookupKey.Tier, lookupKey.Id, clusterKey.Namespace, clusterKey.Cluster)
		}
	}

	return lsPerCluster, nil
}

func findDruidClusterUrls(ctx context.Context, r client.Reader, overrides map[ClusterKey]string) (map[ClusterKey]string, error) {
	rsvcs := &v1.ServiceList{}
	listOpt := client.MatchingLabels(map[string]string{
		"app":       "druid",
		"component": "router",
	})
	if err := r.List(ctx, rsvcs, listOpt); err != nil {
		return nil, err
	}

	clusters := make(map[ClusterKey]string)
	for _, rsvc := range rsvcs.Items {
		key := ClusterKey{
			Namespace: rsvc.Namespace,
			Cluster:   rsvc.Labels["druid_cr"],
		}

		port, found := findFirst(rsvc.Spec.Ports, func(p v1.ServicePort) bool {
			return p.Name == "service-port"
		})
		if !found {
			return nil, fmt.Errorf(`could not find "service-port" of router service %v/%v`, key.Namespace, rsvc.Name)
		}

		url := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", rsvc.Name, rsvc.Namespace, port.Port)

		if override, found := overrides[key]; found {
			url = override
		}
		_, replaced := replace(clusters, key, url)
		if replaced {
			return nil, fmt.Errorf("duplicate router services found for cluster %v/%v", key.Namespace, key.Cluster)
		}
	}

	return clusters, nil
}

func buildClusters(urls map[ClusterKey]string) (map[ClusterKey]*Cluster, error) {
	httpClient := internalhttp.NewHTTPClient(&http.Client{}, &internalhttp.Auth{BasicAuth: internalhttp.BasicAuth{}})
	cls := make(map[ClusterKey]*Cluster)

	for key, url := range urls {
		cl, err := New(url, httpClient)
		if err != nil {
			return nil, err
		}

		cls[key] = cl
	}

	return cls, nil
}

func findFirst[T any](list []T, pred func(T) bool) (T, bool) {
	for _, elem := range list {
		if pred(elem) {
			return elem, true
		}
	}

	return *new(T), false
}

func replace[K comparable, V any](m map[K]V, k K, v V) (V, bool) {
	prev, present := m[k]
	m[k] = v
	return prev, present
}

// SetupWithManager sets up the controller with the Manager.
func (r *DruidLookupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&druidv1alpha1.DruidLookup{}).
		Complete(r)
}
