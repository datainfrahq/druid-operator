package druid

import (
	"context"
	"fmt"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	deletePVCFinalizerName = "deletepvc.finalizers.druid.apache.org"
)

var (
	defaultFinalizers []string
)

func updateFinalizers(ctx context.Context, sdk client.Client, m *v1alpha1.Druid, emitEvents EventEmitter) error {
	desiredFinalizers := m.GetFinalizers()
	additionFinalizers := defaultFinalizers

	desiredFinalizers = RemoveString(desiredFinalizers, deletePVCFinalizerName)
	if !m.Spec.DisablePVCDeletionFinalizer {
		additionFinalizers = append(additionFinalizers, deletePVCFinalizerName)
	}

	for _, finalizer := range additionFinalizers {
		if !ContainsString(desiredFinalizers, finalizer) {
			desiredFinalizers = append(desiredFinalizers, finalizer)
		}
	}

	m.SetFinalizers(desiredFinalizers)
	_, err := writers.Update(ctx, sdk, m, m, emitEvents)
	if err != nil {
		return err
	}

	return nil
}

func executeFinalizers(ctx context.Context, sdk client.Client, m *v1alpha1.Druid, emitEvents EventEmitter) error {
	if m.Spec.DisablePVCDeletionFinalizer == false {
		if err := executePVCFinalizer(ctx, sdk, m, emitEvents); err != nil {
			return err
		}
	}
	return nil
}

/*
executePVCFinalizer will execute a PVC deletion of all Druid's PVCs.
Flow:
 1. Get sts List and PVC List
 2. Range and Delete sts first and then delete pvc. PVC must be deleted after sts termination has been executed
    else pvc finalizer shall block deletion since a pod/sts is referencing it.
 3. Once delete is executed we block program and return.
*/
func executePVCFinalizer(ctx context.Context, sdk client.Client, m *v1alpha1.Druid, emitEvents EventEmitter) error {
	if ContainsString(m.ObjectMeta.Finalizers, deletePVCFinalizerName) {
		pvcLabels := map[string]string{
			"druid_cr": m.Name,
		}

		pvcList, err := readers.List(ctx, sdk, m, pvcLabels, emitEvents, func() objectList { return &v1.PersistentVolumeClaimList{} }, func(listObj runtime.Object) []object {
			items := listObj.(*v1.PersistentVolumeClaimList).Items
			result := make([]object, len(items))
			for i := 0; i < len(items); i++ {
				result[i] = &items[i]
			}
			return result
		})
		if err != nil {
			return err
		}

		stsList, err := readers.List(ctx, sdk, m, makeLabelsForDruid(m.Name), emitEvents, func() objectList { return &appsv1.StatefulSetList{} }, func(listObj runtime.Object) []object {
			items := listObj.(*appsv1.StatefulSetList).Items
			result := make([]object, len(items))
			for i := 0; i < len(items); i++ {
				result[i] = &items[i]
			}
			return result
		})
		if err != nil {
			return err
		}

		msg := fmt.Sprintf("Trigerring finalizer for CR [%s] in namespace [%s]", m.Name, m.Namespace)
		//		sendEvent(sdk, m, v1.EventTypeNormal, DruidFinalizer, msg)
		logger.Info(msg)
		if err := deleteSTSAndPVC(ctx, sdk, m, stsList, pvcList, emitEvents); err != nil {
			return err
		} else {
			msg := fmt.Sprintf("Finalizer success for CR [%s] in namespace [%s]", m.Name, m.Namespace)
			//			sendEvent(sdk, m, v1.EventTypeNormal, DruidFinalizerSuccess, msg)
			logger.Info(msg)
		}

		// remove our finalizer from the list and update it.
		m.ObjectMeta.Finalizers = RemoveString(m.ObjectMeta.Finalizers, deletePVCFinalizerName)

		_, err = writers.Update(ctx, sdk, m, m, emitEvents)
		if err != nil {
			return err
		}

	}
	return nil
}
