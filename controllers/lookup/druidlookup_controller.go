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
	"encoding/json"
	"errors"
	"fmt"
	druidv1alpha1 "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
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
func (r *DruidLookupReconciler) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	logr := log.FromContext(ctx)

	reports := make(map[types.NamespacedName]Report)

	lookupSpecsPerCluster, err := r.FindLookups(ctx, reports)
	if err != nil {
		return ctrl.Result{}, err
	}

	druidClients, nonFatalErrors, fatalErr := r.FindDruidCluster(ctx)
	if fatalErr != nil {
		return ctrl.Result{}, fatalErr
	}
	for _, nonFatalError := range nonFatalErrors {
		logr.Error(nonFatalError, "error occurred while constructing druid client")
	}

	for key, druidClient := range druidClients {
		lookupSpecs := lookupSpecsPerCluster[key]
		if err := druidClient.Reconcile(lookupSpecs, reports); err != nil {
			logr.Error(
				err,
				"could not reconcile lookups for cluster",
				"namespace", key.Namespace,
				"cluster", key.Name,
			)
		}
	}

	for name, report := range reports {
		if err := r.UpdateStatus(ctx, name, report); err != nil {
			logr.Error(
				err,
				"an error occurred while updating lookup resource status",
				"namespace", name.Namespace,
				"name", name.Name,
			)
		}
	}

	for clusterKey, druidClient := range druidClients {
		lookupStatuses, err := druidClient.GetStatus()
		if err != nil {
			logr.Error(
				err,
				"couldn't fetch lookup statues",
				"namespace", clusterKey.Namespace,
				"cluster", clusterKey.Name,
			)
			continue
		}

		for lookupKey, lookupStatus := range lookupStatuses {
			lookupSpec, found := lookupSpecsPerCluster[clusterKey][lookupKey]
			if !found {
				continue
			}

			name := lookupSpec.name

			if err := r.UpdateStatus(ctx, name, &lookupStatus); err != nil {
				logr.Error(
					err,
					"an error occurred while updating lookup resource status",
					"namespace", name.Namespace,
					"name", name.Name,
				)
			}
		}
	}

	return ctrl.Result{RequeueAfter: time.Second * 10}, nil
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

func setIfNotPresent[K comparable, V any](m map[K]V, k K, v V) {
	if _, present := m[k]; !present {
		m[k] = v
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DruidLookupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&druidv1alpha1.DruidLookup{}).
		Complete(r)
}

func (r *DruidLookupReconciler) FindLookups(ctx context.Context, reports map[types.NamespacedName]Report) (LookupsPerCluster, error) {
	lookups := &druidv1alpha1.DruidLookupList{}
	if err := r.List(ctx, lookups); err != nil {
		return nil, err
	}

	lookupSpecsPerCluster := make(LookupsPerCluster)
	for _, lookup := range lookups.Items {
		clusterKey := types.NamespacedName{
			Namespace: lookup.Namespace,
			Name:      lookup.Spec.DruidClusterName,
		}
		lookupKey := LookupKey{
			Tier: lookup.Spec.Tier,
			Id:   lookup.Spec.Id,
		}
		var lookupSpec interface{}
		if err := json.Unmarshal([]byte(lookup.Spec.Spec), &lookupSpec); err != nil {
			setIfNotPresent(
				reports,
				lookup.GetNamespacedName(),
				Report(NewErrorReport(
					fmt.Errorf(
						"lookup resource %v in cluster %v/%v contains invalid spec, should be JSON",
						lookup.Name,
						clusterKey.Namespace,
						clusterKey.Name,
					),
				)),
			)
			continue
		}

		if lookupSpecsPerCluster[clusterKey] == nil {
			lookupSpecsPerCluster[clusterKey] = make(map[LookupKey]Spec)
		}
		ls := lookupSpecsPerCluster[clusterKey]

		if _, replaced := replace(ls, lookupKey, Spec{
			name: lookup.GetNamespacedName(),
			spec: lookupSpec,
		}); replaced {
			setIfNotPresent(
				reports,
				lookup.GetNamespacedName(),
				Report(NewErrorReport(
					fmt.Errorf(
						"resource %v specifies duplicate lookup %v/%v in cluster %v/%v",
						lookup.Name,
						lookupKey.Tier,
						lookupKey.Id,
						clusterKey.Namespace,
						clusterKey.Name,
					),
				)),
			)
			continue
		}
	}

	return lookupSpecsPerCluster, nil
}

func (r *DruidLookupReconciler) FindDruidCluster(ctx context.Context) (map[types.NamespacedName]*DruidClient, []error, error) {
	httpClient := internalhttp.NewHTTPClient(&http.Client{}, &internalhttp.Auth{BasicAuth: internalhttp.BasicAuth{}})
	clusters := make(map[types.NamespacedName]*DruidClient)
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
	if err := r.List(ctx, routerServices, listOpt); err != nil {
		return nil, nil, err
	}

	for _, service := range routerServices.Items {
		key := types.NamespacedName{
			Namespace: service.Namespace,
			Name:      service.Labels["druid_cr"],
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
			nonFatalErrors = append(nonFatalErrors, fmt.Errorf("duplicate router services found for cluster %v/%v", key.Namespace, key.Name))
			continue
		}
	}

	return clusters, nonFatalErrors, nil
}

func (r *DruidLookupReconciler) UpdateStatus(ctx context.Context, name types.NamespacedName, report Report) error {
	lookup := &druidv1alpha1.DruidLookup{}
	err := r.Get(ctx, name, lookup)
	if err != nil {
		return err
	}

	err = report.MergeStatus(&lookup.Status)
	if err != nil {
		return err
	}

	return r.Status().Update(ctx, lookup)
}
