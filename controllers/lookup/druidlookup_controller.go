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
	"errors"
	"fmt"
	druidv1alpha1 "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	"github.com/datainfrahq/druid-operator/controllers/lookup/report"
	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// DruidLookupReconciler reconciles a DruidLookup object
type DruidLookupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	DruidLookupControllerFinalizer = "druidlookup.datainfra.io/finalizer"
)

//+kubebuilder:rbac:groups=druid.apache.org,resources=druidlookups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=druid.apache.org,resources=druidlookups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=druid.apache.org,resources=druidlookups/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *DruidLookupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("reconciling lookup", "namespace", req.Namespace, "name", req.Name)

	lookup := &druidv1alpha1.DruidLookup{}
	if err := r.Get(ctx, req.NamespacedName, lookup); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	druidClients, nonFatalErrors, fatalErr := r.findDruidCluster(ctx)
	if fatalErr != nil {
		return ctrl.Result{}, fatalErr
	}
	for _, nonFatalError := range nonFatalErrors {
		logger.Error(nonFatalError, "error occurred while constructing druid client")
	}

	// examine if lookup is under deletion
	if !lookup.ObjectMeta.DeletionTimestamp.IsZero() {
		// lookup is under deletion
		if err := r.handleDeletingLookup(ctx, druidClients, lookup); err != nil {
			logger.Error(
				err,
				"an error occurred while finalizing lookup resource after handling deleting lookup",
				"namespace", req.NamespacedName.Name,
				"name", req.NamespacedName.Name,
			)
		}

		// do not do any further reconciliation actions
		return ctrl.Result{}, nil
	}

	// lookup is not under deletion
	handleReport := r.handleLookup(ctx, druidClients, lookup)
	if err := r.updateStatus(ctx, req.NamespacedName, handleReport); err != nil {
		logger.Error(
			err,
			"an error occurred while updating lookup resource status after handling lookup",
			"namespace", req.NamespacedName.Name,
			"name", req.NamespacedName.Name,
		)
	}

	statusReport := r.pollLookupStatus(ctx, druidClients, lookup)
	if err := r.updateStatus(ctx, req.NamespacedName, statusReport); err != nil {
		logger.Error(
			err,
			"an error occurred while updating lookup resource status after polling lookup status",
			"namespace", req.NamespacedName.Name,
			"name", req.NamespacedName.Name,
		)
	}

	res := ctrl.Result{}

	if handleReport.ShouldResultInRequeue() || statusReport.ShouldResultInRequeue() {
		res.RequeueAfter = time.Second * 5
	}

	return res, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DruidLookupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&druidv1alpha1.DruidLookup{}).
		Complete(r)
}

func (r *DruidLookupReconciler) handleLookup(ctx context.Context, druidClients map[types.NamespacedName]*DruidClient, lookup *druidv1alpha1.DruidLookup) report.Report {
	// ensure lookup has finalizer registered
	if controllerutil.AddFinalizer(lookup, DruidLookupControllerFinalizer) {
		if err := r.Update(ctx, lookup); err != nil {
			return report.NewErrorReport(err)
		}
	}

	if lookup.ShouldDeleteLastAppliedLookup() {
		druidClient, found := druidClients[types.NamespacedName{Namespace: lookup.Namespace, Name: lookup.Status.LastClusterAppliedIn.Name}]
		if !found {
			return report.NewErrorReport(errors.New(fmt.Sprintf("could not find any druid cluster %s/%s", lookup.Namespace, lookup.Status.LastClusterAppliedIn.Name)))
		}

		if err := druidClient.Delete(lookup.Status.LastTierAppliedIn, lookup.Name); err != nil {
			return report.NewErrorReport(err)
		}
	}

	template, err := lookup.GetTemplateToApply()
	if err != nil {
		return report.NewErrorReport(err)
	}
	if template == nil {
		return report.NewNothingChangedReport()
	}

	druidClient, found := druidClients[types.NamespacedName{Namespace: lookup.Namespace, Name: lookup.Spec.DruidCluster.Name}]
	if !found {
		return report.NewErrorReport(errors.New(fmt.Sprintf("could not find any druid cluster %s/%s", lookup.Namespace, lookup.Spec.DruidCluster.Name)))
	}
	if err := druidClient.Upsert(lookup.Spec.Tier, lookup.Name, template); err != nil {
		return report.NewErrorReport(err)
	}

	return report.NewSuccessReport(lookup.Spec.DruidCluster, lookup.Spec.Tier, lookup.Spec.Template)
}

func (r *DruidLookupReconciler) handleDeletingLookup(ctx context.Context, druidClients map[types.NamespacedName]*DruidClient, lookup *druidv1alpha1.DruidLookup) error {
	if !controllerutil.ContainsFinalizer(lookup, DruidLookupControllerFinalizer) {
		// lookup does not contain our finalizer, therefore, we're already done with this object
		return nil
	}

	// delete last applied lookup
	druidClient, found := druidClients[types.NamespacedName{Namespace: lookup.Namespace, Name: lookup.Status.LastClusterAppliedIn.Name}]
	if !found {
		return errors.New(fmt.Sprintf("could not find any druid cluster %s/%s", lookup.Namespace, lookup.Status.LastClusterAppliedIn.Name))
	}
	if err := druidClient.Delete(lookup.Status.LastTierAppliedIn, lookup.Name); err != nil {
		return err
	}

	controllerutil.RemoveFinalizer(lookup, DruidLookupControllerFinalizer)
	if err := r.Update(ctx, lookup); err != nil {
		return err
	}

	return nil
}

func (r *DruidLookupReconciler) pollLookupStatus(_ context.Context, druidClients map[types.NamespacedName]*DruidClient, lookup *druidv1alpha1.DruidLookup) report.Report {
	druidClient, found := druidClients[types.NamespacedName{Namespace: lookup.Namespace, Name: lookup.Spec.DruidCluster.Name}]
	if !found {
		return report.NewErrorReport(errors.New(fmt.Sprintf("could not find any druid cluster %s/%s", lookup.Namespace, lookup.Spec.DruidCluster.Name)))
	}

	status, err := druidClient.GetStatus(lookup.Spec.Tier, lookup.Name)
	if err != nil {
		return report.NewErrorReport(err)
	}

	return &status
}

func (r *DruidLookupReconciler) findDruidCluster(ctx context.Context) (map[types.NamespacedName]*DruidClient, []error, error) {
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

		if _, overrideExists := overrides[key]; overrideExists {
			continue
		}

		var port *v1.ServicePort = nil
		for _, p := range service.Spec.Ports {
			if p.Name == "service-port" {
				port = &p
			}
		}

		if port == nil {
			nonFatalErrors = append(nonFatalErrors, fmt.Errorf(`could not find "service-port" of router service %v/%v`, key.Namespace, service.Name))
			continue
		}

		url := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", service.Name, service.Namespace, port.Port)

		cluster, err := NewCluster(url, httpClient)
		if err != nil {
			nonFatalErrors = append(nonFatalErrors, errors.Join(fmt.Errorf("could not create druid cluster client for cluster at %v", url), err))
			continue
		}

		if _, ok := clusters[key]; ok {
			nonFatalErrors = append(nonFatalErrors, fmt.Errorf("duplicate router services found for cluster %v/%v", key.Namespace, key.Name))
			continue
		}

		clusters[key] = cluster
	}

	for key, url := range overrides {
		cluster, err := NewCluster(url, httpClient)
		if err != nil {
			nonFatalErrors = append(nonFatalErrors, errors.Join(fmt.Errorf("could not create druid cluster client for cluster at %v", url), err))
			continue
		}

		if _, ok := clusters[key]; ok {
			nonFatalErrors = append(nonFatalErrors, fmt.Errorf("duplicate router services found for cluster %v/%v", key.Namespace, key.Name))
			continue
		}
		clusters[key] = cluster
	}

	return clusters, nonFatalErrors, nil
}

func (r *DruidLookupReconciler) updateStatus(ctx context.Context, name types.NamespacedName, report report.Report) error {
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
