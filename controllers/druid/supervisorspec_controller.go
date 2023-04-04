/*

 */

package druid

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-resty/resty/v2"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	druidv1alpha1 "github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
)

type DruidApiSupervisor struct {
	Id            string          `json:"id"`
	State         string          `json:"state"`
	DetailedState string          `json:"detailedState"`
	Healthy       bool            `json:"healthy"`
	Spec          json.RawMessage `json:"spec"`
	Suspended     bool            `json:"suspended"`
}

type DruidApiSupervisorSpecMinDataSchema struct {
	DataSource string `json:"dataSource"`
}

type DruidApiSupervisorSpecMin struct {
	DataSchema DruidApiSupervisorSpecMinDataSchema `json:"dataSchema"`
}

type SupervisorSpecStateEntry struct {
	Id                   string `json:"id"`
	DruidServiceEndpoint string `json:"druidHost"`
}

const (
	ActionCreate                  = "create"
	ActionUpdate                  = "update"
	ActionDelete                  = "delete"
	SupervisorSpecConfigMap       = "supervisor-specs-controller"
	SupervisorListUrlPattern      = "http://%s:8088/druid/indexer/v1/supervisor"
	SupervisorUrlPattern          = "http://%s:8088/druid/indexer/v1/supervisor/%s"
	SupervisorTerminateUrlPattern = "http://%s:8088/druid/indexer/v1/supervisor/%s/terminate"
	SupervisorResumeUrlPattern    = "http://%s:8088/druid/indexer/v1/supervisor/%s/resume"
	SupervisorSuspendUrlPattern   = "http://%s:8088/druid/indexer/v1/supervisor/%s/suspend"
)

// SupervisorSpecReconciler reconciles a SupervisorSpec object
type SupervisorSpecReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=druid.apache.org,resources=supervisorspecs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=druid.apache.org,resources=supervisorspecs/status,verbs=get;update;patch

// SetupWithManager sets up the controller with the Manager.
func (r *SupervisorSpecReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&druidv1alpha1.SupervisorSpec{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SupervisorSpec object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.6.4/pkg/reconcile
func (r *SupervisorSpecReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithName(string(uuid.NewUUID()))
	_ = log.WithValues("supervisorspec", req.NamespacedName)

	log.Info(fmt.Sprintf("reconciling SupervisorSpec (%s)", req.NamespacedName.String()))

	action := ActionCreate
	supervisorCr := &druidv1alpha1.SupervisorSpec{}
	err := r.Client.Get(ctx, req.NamespacedName, supervisorCr)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "failed to get SupervisorSpec from k8s api")
		}
		action = ActionDelete
	}

	supervisorSpec := supervisorCr.Spec

	requeue := false
	switch action {
	case ActionCreate: // and update
		requeue, err = r.createOrUpdateSupervisorSpec(ctx, log, req, supervisorSpec)
		if err != nil {
			r.updateSyncedStatus(ctx, log, req, false)

			return ctrl.Result{}, err
		}
		if requeue {
			r.updateSyncedStatus(ctx, log, req, false)

			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: LookupReconcileTime(),
			}, nil
		}

		r.updateSyncedStatus(ctx, log, req, true)
	case ActionDelete:
		requeue, err = r.deleteSupervisorSpec(ctx, log, req)
		if err != nil {
			return ctrl.Result{}, err
		}
		if requeue {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: LookupReconcileTime(),
			}, nil
		}
	default:
		log.Error(fmt.Errorf("unexpected action: %s", action), "Error occurred")
	}

	log.Info(fmt.Sprintf("reconciled SupervisorSpec (%s)", req.NamespacedName.String()))

	return ctrl.Result{}, nil
}

func (r *SupervisorSpecReconciler) objectKeyFromStringSlice(input []string, fallbackNamespace string) client.ObjectKey {
	namespace := ""
	name := ""
	if len(input) < 2 {
		namespace = ""
		name = input[0]
	} else {
		namespace = input[0]
		name = input[1]
	}
	if len(namespace) == 0 {
		namespace = fallbackNamespace
	}

	return client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}
}

func (r *SupervisorSpecReconciler) getSupervisorSpecStateEntry(ctx context.Context, log logr.Logger, req ctrl.Request) (*SupervisorSpecStateEntry, error) {
	state := v1.ConfigMap{}
	stateKey := r.objectKeyFromStringSlice([]string{req.Namespace, SupervisorSpecConfigMap}, "")
	err := r.Client.Get(ctx, stateKey, &state)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		} else {
			log.Error(err, fmt.Sprintf("failed to get supervisor spec state configmap %s for reading", stateKey.String()))
			return nil, err
		}
	}

	ssse := &SupervisorSpecStateEntry{}

	entry, ok := state.Data[req.Name]
	if !ok {
		return nil, nil
	}

	err = json.Unmarshal([]byte(entry), ssse)
	if err != nil {
		log.Error(err, "failed to unmarshal supervisor spec state configmap")
		return nil, err
	}

	return ssse, nil
}

func (r *SupervisorSpecReconciler) fetchDruidServicesWithNsList(ctx context.Context, log logr.Logger, clusterRefS string, specNamespace string) ([]string, bool, error) {
	druid := &druidv1alpha1.Druid{}
	clusterRef := strings.Split(clusterRefS, "/")
	if len(clusterRef) < 1 {
		log.Error(fmt.Errorf("clusterRef is empty"), "The cluster reference is invalid")

		return nil, false, nil
	}

	druidObjectKey := r.objectKeyFromStringSlice(clusterRef, specNamespace)
	err := r.Client.Get(ctx, druidObjectKey, druid)
	if err != nil {
		log.Error(err, "failed to get druid from k8s api")
	}

	druidObjectServices := druid.Status.Services
	if len(druidObjectServices) == 0 {
		log.Error(fmt.Errorf("no services in druid spec"), "The services in the status field of the druid spec were empty")

		return nil, true, nil
	}

	druidServices := make([]string, len(druidObjectServices), len(druidObjectServices))
	for i, service := range druidObjectServices {
		druidServices[i] = fmt.Sprintf("%s.%s", service, specNamespace)
	}

	return druidServices, false, nil
}

func (r *SupervisorSpecReconciler) putSupervisorSpecStateEntry(ctx context.Context, log logr.Logger, req ctrl.Request, druidService, id string) error {
	state := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SupervisorSpecConfigMap,
			Namespace: req.Namespace,
		},
	}
	stateKey := r.objectKeyFromStringSlice([]string{req.Namespace, SupervisorSpecConfigMap}, "")
	action := ActionUpdate

	err := r.Client.Get(ctx, stateKey, &state)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, fmt.Sprintf("failed to get supervisor spec state configmap %s for update", stateKey.String()))
			return err
		}

		action = ActionCreate
	}

	ssse := SupervisorSpecStateEntry{
		Id:                   id,
		DruidServiceEndpoint: druidService,
	}

	entry, err := json.Marshal(ssse)
	if err != nil {
		log.Error(err, "failed to marshal supervisor spec state configmap")
		return err
	}

	if state.Data == nil {
		state.Data = map[string]string{}
	}

	state.Data[req.Name] = string(entry)

	if action == ActionCreate {
		err = r.Client.Create(ctx, &state)
	} else {
		err = r.Client.Update(ctx, &state)
	}

	if err != nil {
		log.Error(err, "failed to set supervisor spec state configmap")
		return err
	}

	return nil
}

func (r *SupervisorSpecReconciler) getDruidHostname(ctx context.Context, log logr.Logger, req ctrl.Request, clusterRef string) (*string, error) {
	state, err := r.getSupervisorSpecStateEntry(ctx, log, req)
	if err != nil {
		return nil, err
	}

	druidServices := make([]string, 0)
	hostnames := make([]string, 0)
	if state != nil {
		hostnames = []string{state.DruidServiceEndpoint}
	}

	var requeue bool
	druidServices, requeue, err = r.fetchDruidServicesWithNsList(ctx, log, clusterRef, req.Namespace)
	if err != nil {
		return nil, err
	}

	if requeue || len(druidServices) == 0 {
		return nil, nil
	}

	hostnames = append(hostnames, druidServices...)

	// check spec for whether it needs an update
	rst := resty.New()
	druidResponse := &resty.Response{}
	url := ""
	for _, hostname := range hostnames {
		url = fmt.Sprintf(SupervisorListUrlPattern, hostname)
		druidResponse, err = rst.NewRequest().
			SetContext(ctx).
			SetHeader("Accept", "application/json").
			Get(url)
		if err != nil {
			log.Error(err, fmt.Sprintf("the request could not be successfully executed (%s)", url))
			continue
		}

		if druidResponse.StatusCode() > 399 {
			continue
		}

		return &hostname, nil
	}

	return nil, nil
}

func (r *SupervisorSpecReconciler) getFullDruidSupervisorList(ctx context.Context, log logr.Logger, rst *resty.Client, hostname string) ([]DruidApiSupervisor, error) {
	url := fmt.Sprintf(SupervisorListUrlPattern, hostname)
	response, err := rst.NewRequest().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetQueryParam("state", "true").
		SetQueryParam("full", "true").
		Get(url)
	if err != nil {
		log.Error(err, fmt.Sprintf("the request could not be successfully executed (%s)", url))
		return nil, err
	}

	if response.StatusCode() > 399 {
		return nil, nil
	}

	responseBody := response.Body()
	druidApiSupervisorList := make([]DruidApiSupervisor, 0)

	err = json.Unmarshal(responseBody, &druidApiSupervisorList)
	if err != nil {
		return nil, err
	}

	return druidApiSupervisorList, nil
}

func (r *SupervisorSpecReconciler) compareSpecAndUpdateSyncStatus(ctx context.Context, log logr.Logger, req ctrl.Request, supervisorList []DruidApiSupervisor, k8sApiSupervisorSpec druidv1alpha1.SupervisorSpecSpec) error {
	k8sSpec := map[string]any{}
	err := json.Unmarshal([]byte(k8sApiSupervisorSpec.SupervisorSpec), &k8sSpec)
	if err != nil {
		log.Error(fmt.Errorf("%w: %s", err, k8sApiSupervisorSpec.SupervisorSpec), "failed to unmarshal spec data from k8s api")
	}

	datasourceName := k8sSpec["spec"].(map[string]any)["dataSchema"].(map[string]any)["dataSource"].(string)
	for _, druidApiSupervisorSpec := range supervisorList {
		minimalSpec := &DruidApiSupervisorSpecMin{}
		err := json.Unmarshal(druidApiSupervisorSpec.Spec, minimalSpec)
		if err != nil {
			log.Error(fmt.Errorf("%w: %s", err, string(druidApiSupervisorSpec.Spec)), "failed to unmarshal spec data from druid api into minimal spec object")
			continue
		}

		if minimalSpec.DataSchema.DataSource != datasourceName {
			continue
		}

		druidSpec := map[string]any{}
		err = json.Unmarshal(druidApiSupervisorSpec.Spec, &druidSpec)
		if err != nil {
			log.Error(fmt.Errorf("%w: %s", err, k8sApiSupervisorSpec.SupervisorSpec), "failed to unmarshal spec data from druid api")
			break
		}

		// for some reason druid has the dataSchema, ioConfig and tuningConfig also on the top level
		delete(druidSpec, "dataSchema")
		delete(druidSpec, "tuningConfig")
		delete(druidSpec, "ioConfig")

		if !reflect.DeepEqual(k8sSpec, druidSpec) {
			r.updateSyncedStatus(ctx, log, req, false)
		}

		return nil
	}

	return nil
}

func (r *SupervisorSpecReconciler) doCreateOrUpdateSupervisor(ctx context.Context, log logr.Logger, rst *resty.Client, hostname string, spec string) (*string, error) {
	url := fmt.Sprintf(SupervisorListUrlPattern, hostname)
	druidResponse, err := rst.NewRequest().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(spec).
		Post(url)
	if err != nil {
		log.Error(err, fmt.Sprintf("the request could not be successfully executed (create: %s)", url))
		return nil, err
	}

	v := map[string]string{}
	err = json.Unmarshal(druidResponse.Body(), &v)
	if err != nil {
		log.Error(err, "failed to unmarshal supervisor spec creation response")
		return nil, err
	}
	id := v["id"]

	return &id, nil
}

func (r *SupervisorSpecReconciler) createOrUpdateSupervisorSpec(ctx context.Context, log logr.Logger, req ctrl.Request, k8sApiSupervisorSpec druidv1alpha1.SupervisorSpecSpec) (bool, error) {
	// check spec for whether it needs an update
	rst := resty.New()

	druidHostname, err := r.getDruidHostname(ctx, log, req, k8sApiSupervisorSpec.ClusterRef)
	if err != nil {
		return true, err
	}
	if druidHostname == nil {
		return true, nil
	}

	druidApiSupervisorList := make([]DruidApiSupervisor, 0)
	druidApiSupervisorList, err = r.getFullDruidSupervisorList(ctx, log, rst, *druidHostname)
	if err != nil {
		return true, err
	}

	err = r.compareSpecAndUpdateSyncStatus(ctx, log, req, druidApiSupervisorList, k8sApiSupervisorSpec)
	if err != nil {
		return true, err
	}

	var supervisorId *string
	supervisorId, err = r.doCreateOrUpdateSupervisor(ctx, log, rst, *druidHostname, k8sApiSupervisorSpec.SupervisorSpec)
	if err != nil {
		return true, err
	}

	err = r.putSupervisorSpecStateEntry(ctx, log, req, *druidHostname, *supervisorId)
	if err != nil {
		log.Error(err, "failed to put supervisor spec state configmap")
		return true, nil
	}

	druidSupervisorSuspended, err := r.getSupervisorSuspendedStatus(ctx, log, rst, *druidHostname, *supervisorId)
	if err != nil {
		log.Error(err, "failed to get supervisor status")
		return true, nil
	}

	if druidSupervisorSuspended && !k8sApiSupervisorSpec.Suspend {
		err = r.resumeSupervisor(ctx, log, rst, *druidHostname, *supervisorId)
		if err != nil {
			log.Error(err, "failed to resume supervisor")
			return true, nil
		}
	}

	if !druidSupervisorSuspended && k8sApiSupervisorSpec.Suspend {
		err = r.suspendSupervisor(ctx, log, rst, *druidHostname, *supervisorId)
		if err != nil {
			log.Error(err, "failed to suspend supervisor")
			return true, nil
		}
	}

	return false, nil
}

func (r *SupervisorSpecReconciler) deleteSupervisorSpecStateRef(ctx context.Context, _ logr.Logger, req ctrl.Request) error {
	state := v1.ConfigMap{}
	stateKey := r.objectKeyFromStringSlice([]string{req.Namespace, SupervisorSpecConfigMap}, "")
	err := r.Client.Get(ctx, stateKey, &state)
	if err != nil {
		return err
	}

	delete(state.Data, req.Name)

	return r.Client.Update(ctx, &state)
}

func (r *SupervisorSpecReconciler) deleteSupervisorSpec(ctx context.Context, log logr.Logger, req ctrl.Request) (bool, error) {
	// check spec for whether it needs an update
	rst := resty.New()
	druidResponse := &resty.Response{}
	url := ""

	state, err := r.getSupervisorSpecStateEntry(ctx, log, req)
	if err != nil {
		log.Error(err, "failed to get supervisor spec state configmap for delete action")
		return false, nil
	}
	if state == nil {
		log.Error(fmt.Errorf("state is nil"), "can not delete supervisor spec without reference in state")
		return false, nil
	}

	url = fmt.Sprintf(SupervisorTerminateUrlPattern, state.DruidServiceEndpoint, state.Id)
	druidResponse, err = rst.NewRequest().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		Post(url)

	if err != nil {
		log.Error(err, fmt.Sprintf("the request could not be successfully executed (delete: %s)", url))
		return true, nil
	}

	if druidResponse.StatusCode() > 204 {
		log.Error(fmt.Errorf("received status code %d", druidResponse.StatusCode()), "unexpected status code received")
	}

	err = r.deleteSupervisorSpecStateRef(ctx, log, req)
	if err != nil {
		log.Error(err, "failed to delete supervisor spec state configmap for delete action")
		return false, nil
	}

	return false, nil
}

func (r *SupervisorSpecReconciler) getSupervisorSuspendedStatus(ctx context.Context, log logr.Logger, rst *resty.Client, druidService string, id string) (bool, error) {
	url := fmt.Sprintf(SupervisorUrlPattern, druidService, id)
	druidResponse, err := rst.NewRequest().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		Get(url)
	if err != nil {
		log.Error(err, fmt.Sprintf("the request could not be successfully executed (%s)", url))
		return false, err
	}

	if druidResponse.StatusCode() > 399 {
		return false, fmt.Errorf("unexpected status code: %d", druidResponse.StatusCode())
	}

	res := map[string]any{}
	err = json.Unmarshal(druidResponse.Body(), &res)
	if err != nil {
		log.Error(err, "could not unmarshal druid supervisor spec")
		return false, err
	}

	return res["suspended"].(bool), nil
}

func (r *SupervisorSpecReconciler) resumeSupervisor(ctx context.Context, _ logr.Logger, rst *resty.Client, druidService string, id string) error {
	url := fmt.Sprintf(SupervisorResumeUrlPattern, druidService, id)
	druidResponse, err := rst.NewRequest().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		Post(url)
	if err != nil {
		return err
	}

	// bad request might indicate it is already running: {"error":"[<supervisor-id>] is already running"}
	if druidResponse.StatusCode() > 400 {
		return fmt.Errorf("unexpected status code: %d", druidResponse.StatusCode())
	}

	return nil
}

func (r *SupervisorSpecReconciler) suspendSupervisor(ctx context.Context, _ logr.Logger, rst *resty.Client, druidService string, id string) error {
	url := fmt.Sprintf(SupervisorSuspendUrlPattern, druidService, id)
	druidResponse, err := rst.NewRequest().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		Post(url)
	if err != nil {
		return err
	}

	// bad request might indicate it is already suspended: {"error":"[<supervisor-id>] is already suspended"}
	if druidResponse.StatusCode() > 400 {
		return fmt.Errorf("unexpected status code: %d", druidResponse.StatusCode())
	}

	return nil
}

func (r *SupervisorSpecReconciler) updateSyncedStatus(ctx context.Context, log logr.Logger, req ctrl.Request, synced bool) {
	spec := &druidv1alpha1.SupervisorSpec{}
	err := r.Client.Get(ctx, req.NamespacedName, spec)
	if err != nil {
		log.Error(err, "failed to get SupervisorSpec from k8s api")
		return
	}

	if spec.Status.Synced == fmt.Sprint(synced) {
		return
	}

	spec.Status.Synced = fmt.Sprint(synced)
	err = r.Status().Update(ctx, spec)
	if err != nil {
		log.Error(err, "Failed to update status")
	}
}
