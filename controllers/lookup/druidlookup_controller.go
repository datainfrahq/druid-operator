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
	druidv1alpha1 "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	k8sClient := NewK8sClient(r.Client)

	lookupSpecsPerCluster, err := k8sClient.FindLookups(ctx, reports)
	if err != nil {
		return ctrl.Result{}, err
	}

	druidClients, nonFatalErrors, fatalErr := k8sClient.FindDruidCluster(ctx)
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
		if err := k8sClient.UpdateStatus(ctx, name, report); err != nil {
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

			if err := k8sClient.UpdateStatus(ctx, name, &lookupStatus); err != nil {
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
