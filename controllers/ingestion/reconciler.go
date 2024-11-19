package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"time"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	"github.com/datainfrahq/druid-operator/controllers/druid"
	internalhttp "github.com/datainfrahq/druid-operator/pkg/http"
	"github.com/datainfrahq/operator-runtime/builder"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	DruidIngestionControllerShutDownSuccess    = "DruidIngestionControllerShutDownSuccess"
	DruidIngestionControllerShutDownFail       = "DruidIngestionControllerShutDownFail"
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

	if di.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(di, DruidIngestionControllerFinalizer) {
			controllerutil.AddFinalizer(di, DruidIngestionControllerFinalizer)
			if err := r.Update(ctx, di.DeepCopyObject().(*v1alpha1.DruidIngestion)); err != nil {
				return nil
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(di, DruidIngestionControllerFinalizer) {
			// our finalizer is present, so lets handle any external dependency
			svcName, err := r.getRouterSvcUrl(di.Namespace, di.Spec.DruidClusterName)
			if err != nil {
				return err
			}

			posthttp := internalhttp.NewHTTPClient(
				&http.Client{},
				&internalhttp.Auth{BasicAuth: basicAuth},
			)

			respShutDownTask, err := posthttp.Do(http.MethodPost, getPath(di.Spec.Ingestion.Type, svcName, http.MethodPost, di.Status.TaskId, true), []byte{})
			if err != nil {
				return err
			}
			if respShutDownTask.StatusCode != 200 {
				build.Recorder.GenericEvent(
					di,
					v1.EventTypeWarning,
					fmt.Sprintf("Resp [%s], StatusCode [%d]", string(respShutDownTask.ResponseBody), respShutDownTask.StatusCode),
					DruidIngestionControllerShutDownFail,
				)
			} else {
				build.Recorder.GenericEvent(
					di,
					v1.EventTypeNormal,
					fmt.Sprintf("Resp [%s], StatusCode [%d]", string(respShutDownTask.ResponseBody), respShutDownTask.StatusCode),
					DruidIngestionControllerShutDownSuccess,
				)
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(di, DruidIngestionControllerFinalizer)
			if err := r.Update(ctx, di.DeepCopyObject().(*v1alpha1.DruidIngestion)); err != nil {
				return nil
			}
		}
	}

	return nil
}

// extractDataSource extracts the dataSource from the DruidIngestion spec.
func extractDataSource(di *v1alpha1.DruidIngestion) (string, error) {
	var specData map[string]interface{}
	if err := json.Unmarshal([]byte(di.Spec.Ingestion.Spec), &specData); err != nil {
		return "", err
	}

	spec, ok := specData["spec"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("spec is required in the ingestion spec")
	}

	dataSchema, ok := spec["dataSchema"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("dataSchema is required in the spec")
	}

	dataSource, ok := dataSchema["dataSource"].(string)
	if !ok {
		return "", fmt.Errorf("dataSource is required in the dataSchema")
	}

	return dataSource, nil
}

// UpdateCompaction updates the compaction settings for a Druid data source.
func (r *DruidIngestionReconciler) UpdateCompaction(
	di *v1alpha1.DruidIngestion,
	svcName string,
	dataSource string,
	auth internalhttp.Auth,
) (bool, error) {
	// If there are no compaction settings, return true
	if reflect.DeepEqual(di.Spec.Ingestion.Compaction, v1alpha1.Compaction{}) {
		return true, nil
	}

	postHttp := internalhttp.NewHTTPClient(
		&http.Client{},
		&auth,
	)

	// Convert Compaction struct to JSON
	compactionData, err := json.Marshal(di.Spec.Ingestion.Compaction)
	if err != nil {
		return false, err
	}

	// Convert JSON to map for modification
	var compactionMap map[string]interface{}
	if err := json.Unmarshal(compactionData, &compactionMap); err != nil {
		return false, err
	}

	// Add dataSource to the map
	compactionMap["dataSource"] = dataSource

	// Convert modified map back to JSON
	finalCompactionData, err := json.Marshal(compactionMap)
	if err != nil {
		return false, err
	}

	// Update compaction settings
	respUpdateCompaction, err := postHttp.Do(
		http.MethodPost,
		makePath(svcName, "coordinator", "config", "compaction"),
		[]byte(finalCompactionData),
	)

	if err != nil {
		return false, err
	}

	if respUpdateCompaction.StatusCode == 200 {
		return true, nil
	}

	return false, fmt.Errorf(
		"failed to update compaction, status code: %d, response body: %s",
		respUpdateCompaction.StatusCode, respUpdateCompaction.ResponseBody)
}

// UpdateRules updates the rules for a Druid data source.
func (r *DruidIngestionReconciler) UpdateRules(
	di *v1alpha1.DruidIngestion,
	svcName string,
	dataSource string,
	auth internalhttp.Auth,
) (bool, error) {
	// If there are no rules, return true
	if len(di.Spec.Ingestion.Rules) == 0 {
		return true, nil
	}

	postHttp := internalhttp.NewHTTPClient(
		&http.Client{},
		&auth,
	)

	// Convert Rules slice to JSON
	rulesData, err := json.Marshal(di.Spec.Ingestion.Rules)
	if err != nil {
		return false, err
	}

	// Update rules
	respUpdateRules, err := postHttp.Do(
		http.MethodPost,
		makePath(svcName, "coordinator", "rules", dataSource),
		rulesData,
	)

	if err != nil {
		return false, err
	}

	if respUpdateRules.StatusCode == 200 {
		return true, nil
	}

	return false, fmt.Errorf("failed to update rules, status code: %d, response body: %s", respUpdateRules.StatusCode, respUpdateRules.ResponseBody)
}

func (r *DruidIngestionReconciler) CreateOrUpdate(
	di *v1alpha1.DruidIngestion,
	svcName string,
	build builder.Builder,
	auth internalhttp.Auth,
) (controllerutil.OperationResult, error) {

	dataSource, err := extractDataSource(di)
	if err != nil {
		return controllerutil.OperationResultNone, err
	}

	// check if task id does not exist in status
	if di.Status.TaskId == "" && di.Status.CurrentIngestionSpec == "" {
		// if does not exist create task
		postHttp := internalhttp.NewHTTPClient(
			&http.Client{},
			&auth,
		)

		// Create ingestion task
		respCreateTask, err := postHttp.Do(
			http.MethodPost,
			getPath(di.Spec.Ingestion.Type, svcName, http.MethodPost, "", false),
			[]byte(di.Spec.Ingestion.Spec),
		)

		if err != nil {
			return controllerutil.OperationResultNone, err
		}

		compactionOk, err := r.UpdateCompaction(di, svcName, dataSource, auth)
		if err != nil {
			return controllerutil.OperationResultNone, err
		}

		rulesOk, err := r.UpdateRules(di, svcName, dataSource, auth)
		if err != nil {
			return controllerutil.OperationResultNone, err
		}

		// If the task creation was successful, patch the status with the new task ID.
		if respCreateTask.StatusCode == 200 && compactionOk && rulesOk {
			taskId, err := getTaskIdFromResponse(respCreateTask.ResponseBody)
			if err != nil {
				return controllerutil.OperationResultNone, err
			}
			result, err := r.makePatchDruidIngestionStatus(
				di,
				taskId,
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
		} else {
			// If task creation failed, patch the status to reflect the failure.
			taskId, err := getTaskIdFromResponse(respCreateTask.ResponseBody)
			if err != nil {
				return controllerutil.OperationResultNone, err
			}
			_, err = r.makePatchDruidIngestionStatus(
				di,
				taskId,
				DruidIngestionControllerCreateFail,
				string(respCreateTask.ResponseBody),
				v1.ConditionTrue,
				DruidIngestionControllerCreateFail,
			)
			if err != nil {
				return controllerutil.OperationResultNone, err
			}
			build.Recorder.GenericEvent(
				di,
				v1.EventTypeWarning,
				fmt.Sprintf("Resp [%s], Status", string(respCreateTask.ResponseBody)),
				DruidIngestionControllerCreateFail,
			)
			return controllerutil.OperationResultCreated, nil
		}
	} else {
		// compare the ingestion spec state
		ok, err := druid.IsEqualJson(di.Status.CurrentIngestionSpec, di.Spec.Ingestion.Spec)
		if err != nil {
			return controllerutil.OperationResultNone, err
		}

		if !ok {
			postHttp := internalhttp.NewHTTPClient(
				&http.Client{},
				&auth,
			)

			respUpdateSpec, err := postHttp.Do(
				http.MethodPost,
				getPath(di.Spec.Ingestion.Type, svcName, http.MethodPost, "", false),
				[]byte(di.Spec.Ingestion.Spec),
			)
			if err != nil {
				return controllerutil.OperationResultNone, err
			}

			if respUpdateSpec.StatusCode == 200 {
				// patch status to store the current valid ingestion spec json
				taskId, err := getTaskIdFromResponse(respUpdateSpec.ResponseBody)
				if err != nil {
					return controllerutil.OperationResultNone, err
				}
				result, err := r.makePatchDruidIngestionStatus(
					di,
					taskId,
					DruidIngestionControllerUpdateSuccess,
					string(respUpdateSpec.ResponseBody),
					v1.ConditionTrue,
					DruidIngestionControllerUpdateSuccess,
				)
				if err != nil {
					return controllerutil.OperationResultNone, err
				}
				build.Recorder.GenericEvent(
					di,
					v1.EventTypeNormal,
					fmt.Sprintf("Resp [%s]", string(respUpdateSpec.ResponseBody)),
					DruidIngestionControllerUpdateSuccess,
				)
				build.Recorder.GenericEvent(
					di,
					v1.EventTypeNormal,
					fmt.Sprintf("Resp [%s], Result [%s]", string(respUpdateSpec.ResponseBody), result),
					DruidIngestionControllerPatchStatusSuccess)
			}

		}

		// compare the compaction state
		compactionEqual := reflect.DeepEqual(di.Status.CurrentCompaction, di.Spec.Ingestion.Compaction)

		if !compactionEqual {
			compactionOk, err := r.UpdateCompaction(di, svcName, dataSource, auth)
			if err != nil {
				return controllerutil.OperationResultNone, err
			}

			if compactionOk {
				// patch status to store the current compaction json
				_, err := r.makePatchDruidIngestionStatus(
					di,
					di.Status.TaskId,
					DruidIngestionControllerUpdateSuccess,
					"compaction updated",
					v1.ConditionTrue,
					DruidIngestionControllerUpdateSuccess,
				)
				if err != nil {
					return controllerutil.OperationResultNone, err
				}
				build.Recorder.GenericEvent(
					di,
					v1.EventTypeNormal,
					"compaction updated",
					DruidIngestionControllerUpdateSuccess,
				)
			}
		}

		// compare the rules state
		rulesEqual := reflect.DeepEqual(di.Status.CurrentRules, di.Spec.Ingestion.Rules)

		if !rulesEqual {
			rulesOk, err := r.UpdateRules(di, svcName, dataSource, auth)
			if err != nil {
				return controllerutil.OperationResultNone, err
			}

			if rulesOk {
				// patch status to store the current valid rules json
				_, err := r.makePatchDruidIngestionStatus(
					di,
					di.Status.TaskId,
					DruidIngestionControllerUpdateSuccess,
					"rules updated",
					v1.ConditionTrue,
					DruidIngestionControllerUpdateSuccess,
				)
				if err != nil {
					return controllerutil.OperationResultNone, err
				}
				build.Recorder.GenericEvent(
					di,
					v1.EventTypeNormal,
					"rules updated",
					DruidIngestionControllerUpdateSuccess,
				)
			}
		}

		return controllerutil.OperationResultUpdated, nil

	}
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
		in.Status.CurrentCompaction = di.Spec.Ingestion.Compaction
		in.Status.CurrentRules = di.Spec.Ingestion.Rules
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

func getPath(
	ingestionType v1alpha1.DruidIngestionMethod,
	svcName, httpMethod, taskId string,
	shutDownTask bool) string {

	switch ingestionType {
	case v1alpha1.NativeBatchIndexParallel:
		if httpMethod == http.MethodGet {
			// get task
			return makePath(svcName, "indexer", "task", taskId)
		} else if httpMethod == http.MethodPost && !shutDownTask {
			// create or update task
			return makePath(svcName, "indexer", "task")
		} else if shutDownTask {
			// shutdown task
			return makePath(svcName, "indexer", "task", taskId, "shutdown")
		}
	case v1alpha1.HadoopIndexHadoop:
	case v1alpha1.Kafka:
		if httpMethod == http.MethodGet {
			// get supervisor task
			return makePath(svcName, "indexer", "supervisor", taskId)
		} else if httpMethod == http.MethodPost && !shutDownTask {
			// create or update supervisor task
			return makePath(svcName, "indexer", "supervisor")
		} else if shutDownTask {
			// shut down supervisor
			return makePath(svcName, "indexer", "supervisor", taskId, "shutdown")
		}
	case v1alpha1.Kinesis:
	case v1alpha1.QueryControllerSQL:
	default:
		return ""
	}

	return ""
}

// makePath constructs the appropriate path for the specified Druid API.
func makePath(baseURL, componentType, apiType string, additionalPaths ...string) string {
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

type taskHolder struct {
	Task string `json:"task"` // tasks
	ID   string `json:"id"`   // supervisor
}

func getTaskIdFromResponse(resp string) (string, error) {
	var task taskHolder
	if err := json.Unmarshal([]byte(resp), &task); err != nil {
		return "", err
	}

	// check both fields and return the appropriate value
	// tasks use different field names than supervisors
	if task.Task != "" {
		return task.Task, nil
	}
	if task.ID != "" {
		return task.ID, nil
	}

	return "", errors.New("task id not found")
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

	newName := "http://" + svcName + "." + namespace + ".svc.cluster.local:" + DruidRouterPort

	return newName, nil
}

func (r *DruidIngestionReconciler) getAuthCreds(ctx context.Context, di *v1alpha1.DruidIngestion) (internalhttp.BasicAuth, error) {
	druid := v1alpha1.Druid{}
	// check if the mentioned druid cluster exists
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: di.Namespace,
		Name:      di.Spec.DruidClusterName,
	},
		&druid,
	); err != nil {
		return internalhttp.BasicAuth{}, err
	}
	// check if the mentioned secret exists
	if di.Spec.Auth != (v1alpha1.Auth{}) {
		secret := v1.Secret{}
		if err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: di.Spec.Auth.SecretRef.Namespace,
			Name:      di.Spec.Auth.SecretRef.Name,
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
