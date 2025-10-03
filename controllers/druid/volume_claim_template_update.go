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

// VolumeClaimTemplateChanges tracks what changed in VolumeClaimTemplates
type VolumeClaimTemplateChanges struct {
	HasChanges                  bool
	AnnotationChanges           map[string]bool // VCT name -> has annotation changes
	VolumeAttributeClassChanges map[string]bool // VCT name -> has volumeAttributeClassName changes
	Details                     string
}

// patchStatefulSetVolumeClaimTemplates handles updates to VolumeClaimTemplates that require StatefulSet recreation
// This includes both annotation and volumeAttributeClassName updates
// Returns (deleted bool, error) - deleted indicates if the StatefulSet was deleted in this call
func patchStatefulSetVolumeClaimTemplates(ctx context.Context, sdk client.Client, m *v1alpha1.Druid,
	nodeSpec *v1alpha1.DruidNodeSpec, emitEvent EventEmitter, nodeSpecUniqueStr string) (bool, error) {

	// Skip if PVC updates are disabled
	if m.Spec.DisablePVCUpdates {
		return false, nil
	}

	// Only process StatefulSets
	if nodeSpec.Kind == "Deployment" {
		return false, nil
	}

	// Get the existing StatefulSet
	sts, err := readers.Get(ctx, sdk, nodeSpecUniqueStr, m, func() object { return &appsv1.StatefulSet{} }, emitEvent)
	if err != nil {
		// StatefulSet doesn't exist yet, will be created with proper configuration
		return false, nil
	}

	statefulSet := sts.(*appsv1.StatefulSet)

	// Check for VolumeClaimTemplate changes
	changes := detectVolumeClaimTemplateChanges(statefulSet, nodeSpec)
	fmt.Println(changes)
	if !changes.HasChanges {
		return false, nil
	}

	// Check if PVCs already have the desired configuration
	pvcAlreadyUpdated, err := checkPVCsAlreadyUpdated(ctx, sdk, statefulSet, nodeSpec, m, emitEvent, nodeSpecUniqueStr)
	if err != nil {
		return false, err
	}
	if pvcAlreadyUpdated {
		// PVCs already updated from a previous reconcile
		return false, nil
	}

	// Don't proceed unless all statefulsets are up and running
	if err := waitForAllStatefulSetsReady(ctx, sdk, m, emitEvent); err != nil {
		return false, nil
	}

	// Emit event for change detection
	msg := fmt.Sprintf("Detected VolumeClaimTemplate changes for StatefulSet [%s] in Namespace [%s]: %s",
		statefulSet.Name, statefulSet.Namespace, changes.Details)
	emitEvent.EmitEventGeneric(m, "VolumeClaimTemplateChangeDetected", msg, nil)

	// Delete the StatefulSet with cascade=orphan
	msg = fmt.Sprintf("Deleting StatefulSet [%s] with cascade=orphan to apply VolumeClaimTemplate changes", statefulSet.Name)
	emitEvent.EmitEventGeneric(m, "StatefulSetOrphanedForVCTChanges", msg, nil)

	if err := writers.Delete(ctx, sdk, m, statefulSet, emitEvent, client.PropagationPolicy(metav1.DeletePropagationOrphan)); err != nil {
		return false, err
	}

	msg = fmt.Sprintf("StatefulSet [%s] successfully deleted with cascade=orphan", statefulSet.Name)
	emitEvent.EmitEventGeneric(m, "StatefulSetOrphanedForVCTChanges", msg, nil)

	// Update PVCs with new configuration
	if err := patchPVCs(ctx, sdk, statefulSet, nodeSpec, m, emitEvent, nodeSpecUniqueStr, changes); err != nil {
		return false, err
	}

	return true, nil
}

// detectVolumeClaimTemplateChanges detects all changes in VolumeClaimTemplates
func detectVolumeClaimTemplateChanges(sts *appsv1.StatefulSet, nodeSpec *v1alpha1.DruidNodeSpec) VolumeClaimTemplateChanges {
	changes := VolumeClaimTemplateChanges{
		HasChanges:                  false,
		AnnotationChanges:           make(map[string]bool),
		VolumeAttributeClassChanges: make(map[string]bool),
	}

	fmt.Println(sts.Name)
	var changeDetails []string

	// Create maps of current VCT configuration
	currentVCTMap := make(map[string]v1.PersistentVolumeClaim)
	for _, vct := range sts.Spec.VolumeClaimTemplates {
		currentVCTMap[vct.Name] = vct
	}

	// Check each desired VCT for changes
	for _, desiredVCT := range nodeSpec.VolumeClaimTemplates {
		currentVCT, exists := currentVCTMap[desiredVCT.Name]
		if !exists {
			// New VCT, skip
			continue
		}

		// Check annotation changes
		if !reflect.DeepEqual(currentVCT.ObjectMeta.Annotations, desiredVCT.ObjectMeta.Annotations) {
			changes.AnnotationChanges[desiredVCT.Name] = true
			changeDetails = append(changeDetails, fmt.Sprintf("VCT %s: annotations changed", desiredVCT.Name))
			changes.HasChanges = true
		}

		// Check volumeAttributeClassName changes
		currentVolAttrClass := ""
		if currentVCT.Spec.VolumeAttributesClassName != nil {
			currentVolAttrClass = *currentVCT.Spec.VolumeAttributesClassName
		}
		desiredVolAttrClass := ""
		if desiredVCT.Spec.VolumeAttributesClassName != nil {
			desiredVolAttrClass = *desiredVCT.Spec.VolumeAttributesClassName
		}

		if currentVolAttrClass != desiredVolAttrClass {
			changes.VolumeAttributeClassChanges[desiredVCT.Name] = true
			changeDetails = append(changeDetails, fmt.Sprintf("VCT %s: volumeAttributeClassName changed from '%s' to '%s'",
				desiredVCT.Name, currentVolAttrClass, desiredVolAttrClass))
			changes.HasChanges = true
		}
	}

	if len(changeDetails) > 0 {
		changes.Details = strings.Join(changeDetails, "; ")
	}

	return changes
}

// patchPVCs patches existing PVCs with new configuration
func patchPVCs(ctx context.Context, sdk client.Client, sts *appsv1.StatefulSet,
	nodeSpec *v1alpha1.DruidNodeSpec, m *v1alpha1.Druid, emitEvent EventEmitter,
	nodeSpecUniqueStr string, changes VolumeClaimTemplateChanges) error {

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

	// Create map of desired VCT configurations
	desiredVCTMap := make(map[string]v1.PersistentVolumeClaim)
	for _, vct := range nodeSpec.VolumeClaimTemplates {
		desiredVCTMap[vct.Name] = vct
	}

	// Patch each PVC with new configuration
	for _, pvcObj := range pvcList {
		pvc := pvcObj.(*v1.PersistentVolumeClaim)

		// Determine which VolumeClaimTemplate this PVC belongs to
		vctName := extractVCTNameFromPVC(pvc.Name, sts.Name)
		if vctName == "" {
			continue
		}

		desiredVCT, exists := desiredVCTMap[vctName]
		if !exists {
			continue
		}

		// Check if this PVC needs updates
		needsPatch := false
		updateMessages := []string{}

		// Check annotations
		if changes.AnnotationChanges[vctName] {
			if !reflect.DeepEqual(pvc.ObjectMeta.Annotations, desiredVCT.ObjectMeta.Annotations) {
				needsPatch = true
				updateMessages = append(updateMessages, "annotations")
			}
		}

		// Check volumeAttributeClassName
		if changes.VolumeAttributeClassChanges[vctName] {
			currentVolAttrClass := ""
			if pvc.Spec.VolumeAttributesClassName != nil {
				currentVolAttrClass = *pvc.Spec.VolumeAttributesClassName
			}
			desiredVolAttrClass := ""
			if desiredVCT.Spec.VolumeAttributesClassName != nil {
				desiredVolAttrClass = *desiredVCT.Spec.VolumeAttributesClassName
			}

			if currentVolAttrClass != desiredVolAttrClass {
				needsPatch = true
				updateMessages = append(updateMessages, fmt.Sprintf("volumeAttributeClassName to '%s'", desiredVolAttrClass))
			}
		}

		if !needsPatch {
			continue
		}

		// Create patch
		pvcCopy := pvc.DeepCopy()
		patch := client.MergeFrom(pvcCopy)

		// Apply annotation updates
		if changes.AnnotationChanges[vctName] {
			if pvc.ObjectMeta.Annotations == nil {
				pvc.ObjectMeta.Annotations = make(map[string]string)
			}

			// Apply desired annotations
			for key, value := range desiredVCT.ObjectMeta.Annotations {
				pvc.ObjectMeta.Annotations[key] = value
			}

			// Remove annotations not in desired state
			for key := range pvc.ObjectMeta.Annotations {
				if _, exists := desiredVCT.ObjectMeta.Annotations[key]; !exists {
					delete(pvc.ObjectMeta.Annotations, key)
				}
			}
		}

		// Apply volumeAttributeClassName update
		if changes.VolumeAttributeClassChanges[vctName] {
			pvc.Spec.VolumeAttributesClassName = desiredVCT.Spec.VolumeAttributesClassName
		}

		// Patch the PVC
		if err := writers.Patch(ctx, sdk, m, pvc, false, patch, emitEvent); err != nil {
			return err
		}

		msg := fmt.Sprintf("PVC [%s] successfully patched with: %s", pvc.Name, strings.Join(updateMessages, ", "))
		emitEvent.EmitEventGeneric(m, "PVCUpdated", msg, nil)
	}

	return nil
}

// checkPVCsAlreadyUpdated checks if PVCs already have the desired configuration
func checkPVCsAlreadyUpdated(ctx context.Context, sdk client.Client, sts *appsv1.StatefulSet,
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

	// Create map of desired VCT configurations
	desiredVCTMap := make(map[string]v1.PersistentVolumeClaim)
	for _, vct := range nodeSpec.VolumeClaimTemplates {
		desiredVCTMap[vct.Name] = vct
	}

	// Check each PVC
	for _, pvcObj := range pvcList {
		pvc := pvcObj.(*v1.PersistentVolumeClaim)

		// Determine which VolumeClaimTemplate this PVC belongs to
		vctName := extractVCTNameFromPVC(pvc.Name, sts.Name)
		if vctName == "" {
			continue
		}

		desiredVCT, exists := desiredVCTMap[vctName]
		if !exists {
			continue
		}

		// Check annotations
		if !reflect.DeepEqual(pvc.ObjectMeta.Annotations, desiredVCT.ObjectMeta.Annotations) {
			return false, nil
		}

		// Check volumeAttributeClassName
		currentVolAttrClass := ""
		if pvc.Spec.VolumeAttributesClassName != nil {
			currentVolAttrClass = *pvc.Spec.VolumeAttributesClassName
		}
		desiredVolAttrClass := ""
		if desiredVCT.Spec.VolumeAttributesClassName != nil {
			desiredVolAttrClass = *desiredVCT.Spec.VolumeAttributesClassName
		}

		if currentVolAttrClass != desiredVolAttrClass {
			return false, nil
		}
	}

	// All PVCs have the desired configuration
	return true, nil
}

// waitForAllStatefulSetsReady waits for all StatefulSets to be ready
func waitForAllStatefulSetsReady(ctx context.Context, sdk client.Client, m *v1alpha1.Druid, emitEvent EventEmitter) error {
	getSTSList, err := readers.List(ctx, sdk, m, makeLabelsForDruid(m), emitEvent, func() objectList { return &appsv1.StatefulSetList{} }, func(listObj runtime.Object) []object {
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

	for _, sts := range getSTSList {
		if sts.(*appsv1.StatefulSet).Status.Replicas != sts.(*appsv1.StatefulSet).Status.ReadyReplicas {
			return fmt.Errorf("not all StatefulSets are ready")
		}
	}

	return nil
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
