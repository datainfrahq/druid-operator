package druid

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// patchStatefulSetVolumeClaimTemplateAnnotations handles annotation updates for StatefulSet VolumeClaimTemplates
// Returns (deleted bool, error) - deleted indicates if the StatefulSet was deleted in this call
func patchStatefulSetVolumeClaimTemplateAnnotations(ctx context.Context, sdk client.Client, m *v1alpha1.Druid,
	nodeSpec *v1alpha1.DruidNodeSpec, emitEvent EventEmitter, nodeSpecUniqueStr string) (bool, error) {

	// Skip if PVC annotation update is disabled
	if m.Spec.DisablePVCAnnotationUpdate {
		return false, nil
	}

	// Only process StatefulSets (default kind is StatefulSet when not specified or anything other than "Deployment")
	if nodeSpec.Kind == "Deployment" {
		return false, nil
	}

	// Get the existing StatefulSet
	sts, err := readers.Get(ctx, sdk, nodeSpecUniqueStr, m, func() object { return &appsv1.StatefulSet{} }, emitEvent)
	if err != nil {
		// StatefulSet doesn't exist yet, will be created with proper annotations
		// This is expected after we delete it for recreation
		return false, nil
	}

	statefulSet := sts.(*appsv1.StatefulSet)

	// Check if annotation changes are needed
	annotationChangesNeeded, annotationDetails := detectAnnotationChanges(statefulSet, nodeSpec)
	if !annotationChangesNeeded {
		return false, nil
	}

	// Before proceeding, check if PVCs already have the desired annotations
	// This prevents re-processing after StatefulSet deletion
	pvcAlreadyUpdated, err := checkPVCAnnotationsAlreadyUpdated(ctx, sdk, statefulSet, nodeSpec, m, emitEvent, nodeSpecUniqueStr)
	if err != nil {
		return false, err
	}
	if pvcAlreadyUpdated {
		// PVCs already have the desired annotations, likely from a previous reconcile
		// The StatefulSet will be recreated with correct annotations by sdkCreateOrUpdateAsNeeded
		return false, nil
	}

	// Don't proceed unless all statefulsets are up and running
	getSTSList, err := readers.List(ctx, sdk, m, makeLabelsForDruid(m), emitEvent, func() objectList { return &appsv1.StatefulSetList{} }, func(listObj runtime.Object) []object {
		items := listObj.(*appsv1.StatefulSetList).Items
		result := make([]object, len(items))
		for i := 0; i < len(items); i++ {
			result[i] = &items[i]
		}
		return result
	})
	if err != nil {
		return false, nil
	}

	for _, sts := range getSTSList {
		if sts.(*appsv1.StatefulSet).Status.Replicas != sts.(*appsv1.StatefulSet).Status.ReadyReplicas {
			return false, nil
		}
	}

	// Emit event for annotation change detection
	msg := fmt.Sprintf("Detected annotation changes in VolumeClaimTemplates for StatefulSet [%s] in Namespace [%s]: %s",
		statefulSet.Name, statefulSet.Namespace, annotationDetails)
	emitEvent.EmitEventGeneric(m, string(druidPvcAnnotationChangeDetected), msg, nil)

	// First, delete the StatefulSet with cascade=orphan (similar to volume expansion)
	msg = fmt.Sprintf("Deleting StatefulSet [%s] with cascade=orphan to apply annotation changes", statefulSet.Name)
	emitEvent.EmitEventGeneric(m, string(druidStsOrphanedForAnnotations), msg, nil)

	if err := writers.Delete(ctx, sdk, m, statefulSet, emitEvent, client.PropagationPolicy(metav1.DeletePropagationOrphan)); err != nil {
		return false, err
	}

	msg = fmt.Sprintf("StatefulSet [%s] successfully deleted with cascade=orphan for annotation updates", statefulSet.Name)
	emitEvent.EmitEventGeneric(m, string(druidStsOrphanedForAnnotations), msg, nil)

	// Then update PVC annotations after deletion (similar to volume expansion)
	if err := patchPVCAnnotations(ctx, sdk, statefulSet, nodeSpec, m, emitEvent, nodeSpecUniqueStr); err != nil {
		return false, err
	}

	// Return true to indicate StatefulSet was deleted
	return true, nil
}

// detectAnnotationChanges compares current and desired annotations for VolumeClaimTemplates
func detectAnnotationChanges(sts *appsv1.StatefulSet, nodeSpec *v1alpha1.DruidNodeSpec) (bool, string) {
	var changeDetails []string

	// Create a map of current VCT annotations by name
	currentVCTAnnotations := make(map[string]map[string]string)
	for _, vct := range sts.Spec.VolumeClaimTemplates {
		currentVCTAnnotations[vct.Name] = vct.ObjectMeta.Annotations
	}

	// Check each desired VCT for annotation changes
	for _, desiredVCT := range nodeSpec.VolumeClaimTemplates {
		currentAnnotations, exists := currentVCTAnnotations[desiredVCT.Name]

		// If VCT doesn't exist in current StatefulSet, skip (it's a new VCT)
		if !exists {
			continue
		}

		// Compare annotations
		if !reflect.DeepEqual(currentAnnotations, desiredVCT.ObjectMeta.Annotations) {
			changeDetails = append(changeDetails, fmt.Sprintf("VCT %s: annotations changed", desiredVCT.Name))
		}
	}

	if len(changeDetails) > 0 {
		return true, strings.Join(changeDetails, "; ")
	}

	return false, ""
}

// patchPVCAnnotations patches existing PVCs with new annotations from VolumeClaimTemplates
func patchPVCAnnotations(ctx context.Context, sdk client.Client, sts *appsv1.StatefulSet,
	nodeSpec *v1alpha1.DruidNodeSpec, m *v1alpha1.Druid, emitEvent EventEmitter, nodeSpecUniqueStr string) error {

	// Get PVCs for this StatefulSet
	pvcLabels := map[string]string{
		"nodeSpecUniqueStr": nodeSpecUniqueStr,
	}

	pvcList, err := readers.List(ctx, sdk, m, pvcLabels, emitEvent, func() objectList { return &v1.PersistentVolumeClaimList{} }, func(listObj runtime.Object) []object {
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

	// Create a map of desired annotations by VCT name
	desiredAnnotationsByVCT := make(map[string]map[string]string)
	for _, vct := range nodeSpec.VolumeClaimTemplates {
		desiredAnnotationsByVCT[vct.Name] = vct.ObjectMeta.Annotations
	}

	// Patch each PVC with new annotations
	for _, pvcObj := range pvcList {
		pvc := pvcObj.(*v1.PersistentVolumeClaim)

		// Determine which VolumeClaimTemplate this PVC belongs to
		vctName := extractVCTNameFromPVC(pvc.Name, sts.Name)
		if vctName == "" {
			continue
		}

		desiredAnnotations, exists := desiredAnnotationsByVCT[vctName]
		if !exists {
			continue
		}

		// Check if annotations need updating
		if reflect.DeepEqual(pvc.ObjectMeta.Annotations, desiredAnnotations) {
			continue
		}

		// Create patch for annotations
		pvcCopy := pvc.DeepCopy()
		patch := client.MergeFrom(pvcCopy)

		// Update or set annotations
		if pvc.ObjectMeta.Annotations == nil {
			pvc.ObjectMeta.Annotations = make(map[string]string)
		}

		// Apply desired annotations
		for key, value := range desiredAnnotations {
			pvc.ObjectMeta.Annotations[key] = value
		}

		// Remove annotations that are not in desired state
		for key := range pvc.ObjectMeta.Annotations {
			if _, exists := desiredAnnotations[key]; !exists {
				delete(pvc.ObjectMeta.Annotations, key)
			}
		}

		// Patch the PVC
		if err := writers.Patch(ctx, sdk, m, pvc, false, patch, emitEvent); err != nil {
			return err
		}

		msg := fmt.Sprintf("PVC [%s] successfully patched with updated annotations", pvc.Name)
		emitEvent.EmitEventGeneric(m, string(druidPvcAnnotationsUpdated), msg, nil)
	}

	return nil
}

// checkPVCAnnotationsAlreadyUpdated checks if PVCs already have the desired annotations
func checkPVCAnnotationsAlreadyUpdated(ctx context.Context, sdk client.Client, sts *appsv1.StatefulSet,
	nodeSpec *v1alpha1.DruidNodeSpec, m *v1alpha1.Druid, emitEvent EventEmitter, nodeSpecUniqueStr string) (bool, error) {

	// Get PVCs for this StatefulSet
	pvcLabels := map[string]string{
		"nodeSpecUniqueStr": nodeSpecUniqueStr,
	}

	pvcList, err := readers.List(ctx, sdk, m, pvcLabels, emitEvent, func() objectList { return &v1.PersistentVolumeClaimList{} }, func(listObj runtime.Object) []object {
		items := listObj.(*v1.PersistentVolumeClaimList).Items
		result := make([]object, len(items))
		for i := 0; i < len(items); i++ {
			result[i] = &items[i]
		}
		return result
	})
	if err != nil {
		return false, err
	}

	// Create a map of desired annotations by VCT name
	desiredAnnotationsByVCT := make(map[string]map[string]string)
	for _, vct := range nodeSpec.VolumeClaimTemplates {
		desiredAnnotationsByVCT[vct.Name] = vct.ObjectMeta.Annotations
	}

	// Check each PVC to see if it already has the desired annotations
	for _, pvcObj := range pvcList {
		pvc := pvcObj.(*v1.PersistentVolumeClaim)

		// Determine which VolumeClaimTemplate this PVC belongs to
		vctName := extractVCTNameFromPVC(pvc.Name, sts.Name)
		if vctName == "" {
			continue
		}

		desiredAnnotations, exists := desiredAnnotationsByVCT[vctName]
		if !exists {
			continue
		}

		// If any PVC doesn't have the desired annotations, return false
		if !reflect.DeepEqual(pvc.ObjectMeta.Annotations, desiredAnnotations) {
			return false, nil
		}
	}

	// All PVCs have the desired annotations
	return true, nil
}

// extractVCTNameFromPVC extracts the VolumeClaimTemplate name from a PVC name
// PVC naming format: {vctName}-{statefulSetName}-{ordinal}
func extractVCTNameFromPVC(pvcName, stsName string) string {
	// Remove the StatefulSet name and ordinal suffix
	suffix := fmt.Sprintf("-%s-", stsName)
	idx := strings.LastIndex(pvcName, suffix)
	if idx == -1 {
		return ""
	}
	return pvcName[:idx]
}
