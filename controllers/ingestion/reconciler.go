package ingestion

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	internalhttp "github.com/datainfrahq/druid-operator/controllers/ingestion/http"
	"github.com/datainfrahq/operator-runtime/builder"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	DruidIngestionControllerCreateSuccess      = "DruidIngestionControllerCreateSuccess"
	DruidIngestionControllerCreateFail         = "DruidIngestionControllerCreateFail"
	DruidIngestionControllerGetSuccess         = "DruidIngestionControllerGetSuccess"
	DruidIngestionControllerGetFail            = "DruidIngestionControllerGetFail"
	DruidIngestionControllerUpdateSuccess      = "DruidIngestionControllerUpdateSuccess"
	DruidIngestionControllerUpdateFail         = "DruidIngestionControllerUpdateFail"
	DruidIngestionControllerDeleteSuccess      = "DruidIngestionControllerDeleteSuccess"
	DruidIngestionControllerDeleteFail         = "DruidIngestionControllerDeleteFail"
	DruidIngestionControllerPatchStatusSuccess = "DruidIngestionControllerPatchStatusSuccess"
	DruidIngestionControllerPatchStatusFail    = "DruidIngestionControllerPatchStatusFail"
	DruidIngestionControllerFinalizer          = "druidingestion.datainfra.io/finalizer"
)

const (
	OperatorUserName = "OperatorUserName"
	OperatorPassword = "OperatorPassword"
)
const (
	DruidRouterPort = "8088"
)

func (r *DruidIngestionReconciler) do(ctx context.Context, di *v1alpha1.DruidIngestion) error {

	basicAuth, err := r.getAuthCreds(ctx, di)
	if err != nil {
		return err
	}

	svcName, err := r.getRouterSvcUrl(di.Namespace, di.Spec.DruidClusterName)
	if err != nil {
		return err
	}

	build := builder.NewBuilder(
		builder.ToNewBuilderRecorder(builder.BuilderRecorder{Recorder: r.Recorder, ControllerName: "DruidIngestionController"}),
	)

	_, err = r.CreateOrUpdate(di, svcName, *build, internalhttp.Auth{BasicAuth: basicAuth})
	if err != nil {
		return err
	}

	return nil
}

func (r *DruidIngestionReconciler) CreateOrUpdate(
	di *v1alpha1.DruidIngestion,
	svcName string,
	build builder.Builder,
	auth internalhttp.Auth,
) (controllerutil.OperationResult, error) {

	// check status if task id exists
	if di.Status.TaskId == "" {
		// if does not exist create task
		postHttp := internalhttp.NewHTTPClient(
			http.MethodPost,
			makeRouterCreateUpdateTaskPath(svcName),
			http.Client{},
			[]byte(di.Spec.Ingestion.Spec),
			auth,
		)

		respCreateTask, err := postHttp.Do()
		if err != nil {
			return controllerutil.OperationResultNone, nil
		}

		// if success patch status
		if respCreateTask.StatusCode == 200 {
			result, err := r.makePatchDruidIngestionStatus(
				di,
				respCreateTask.ResponseBody,
				DruidIngestionControllerCreateSuccess,
				string(respCreateTask.ResponseBody),
				v1.ConditionTrue,
				DruidIngestionControllerCreateSuccess,
			)
			if err != nil {
				return controllerutil.OperationResultNone, err
			}
			build.Recorder.GenericEvent(
				di,
				v1.EventTypeNormal,
				fmt.Sprintf("Resp [%s]", string(respCreateTask.ResponseBody)),
				DruidIngestionControllerCreateSuccess,
			)
			build.Recorder.GenericEvent(
				di,
				v1.EventTypeNormal,
				fmt.Sprintf("Resp [%s], Result [%s]", string(respCreateTask.ResponseBody), result),
				DruidIngestionControllerPatchStatusSuccess)
			return controllerutil.OperationResultCreated, nil
		}
	}

	return controllerutil.OperationResultNone, nil
}

func (r *DruidIngestionReconciler) makePatchDruidIngestionStatus(
	di *v1alpha1.DruidIngestion,
	taskId string,
	msg string,
	reason string,
	status v1.ConditionStatus,
	diConditionType string,

) (controllerutil.OperationResult, error) {

	if _, _, err := patchStatus(context.Background(), r.Client, di, func(obj client.Object) client.Object {
		in := obj.(*v1alpha1.DruidIngestion)
		in.Status.CurrentIngestionSpec = di.Spec.Ingestion.Spec
		in.Status.TaskId = taskId
		in.Status.LastUpdateTime = metav1.Time{Time: time.Now()}
		in.Status.Message = msg
		in.Status.Reason = reason
		in.Status.Status = status
		in.Status.Type = diConditionType
		return in
	}); err != nil {
		return controllerutil.OperationResultNone, err
	}

	return controllerutil.OperationResultUpdatedStatusOnly, nil
}

func makeRouterCreateUpdateTaskPath(svcName string) string {
	return svcName + "/druid/indexer/v1/task"
}

func makeRouterGetTaskPath(svcName, taskId string) string {
	return svcName + "/druid/indexer/v1/task/" + taskId
}

func (r *DruidIngestionReconciler) getRouterSvcUrl(namespace, druidClusterName string) (string, error) {
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{
			"druid_cr":  druidClusterName,
			"component": "router",
		}),
	}
	svcList := &v1.ServiceList{}
	if err := r.Client.List(context.Background(), svcList, listOpts...); err != nil {
		return "", err
	}
	var svcName string

	for range svcList.Items {
		svcName = svcList.Items[0].Name
	}

	if svcName == "" {
		return "", errors.New("router svc discovery fail")
	}
	_ = "http://" + svcName + "." + namespace + ".svc.cluster.local:" + DruidRouterPort

	return "http://localhost:8088", nil
}

func (r *DruidIngestionReconciler) getAuthCreds(ctx context.Context, di *v1alpha1.DruidIngestion) (internalhttp.BasicAuth, error) {
	druid := v1alpha1.Druid{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: di.Namespace,
		Name:      di.Spec.DruidClusterName,
	},
		&druid,
	); err != nil {
		return internalhttp.BasicAuth{}, err
	}

	if druid.Spec.Auth != (v1alpha1.Auth{}) {
		secret := v1.Secret{}
		if err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: druid.Spec.Auth.SecretRef.Namespace,
			Name:      druid.Spec.Auth.SecretRef.Name,
		},
			&secret,
		); err != nil {
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

type VerbType string

type (
	TransformStatusFunc func(obj client.Object) client.Object
)

const (
	VerbPatched   VerbType = "Patched"
	VerbUnchanged VerbType = "Unchanged"
)

func patchStatus(ctx context.Context, c client.Client, obj client.Object, transform TransformStatusFunc, opts ...client.SubResourcePatchOption) (client.Object, VerbType, error) {
	key := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
	err := c.Get(ctx, key, obj)
	if err != nil {
		return nil, VerbUnchanged, err
	}

	// The body of the request was in an unknown format -
	// accepted media types include:
	//   - application/json-patch+json,
	//   - application/merge-patch+json,
	//   - application/apply-patch+yaml
	patch := client.MergeFrom(obj)
	obj = transform(obj.DeepCopyObject().(client.Object))
	err = c.Status().Patch(ctx, obj, patch, opts...)
	if err != nil {
		return nil, VerbUnchanged, err
	}
	return obj, VerbPatched, nil
}
