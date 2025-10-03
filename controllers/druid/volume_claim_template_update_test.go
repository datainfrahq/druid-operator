package druid

import (
	"context"
	"testing"

	"github.com/datainfrahq/druid-operator/apis/druid/v1alpha1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mockEventEmitter provides a no-op implementation of EventEmitter for testing
type mockEventEmitter struct{}

func (m *mockEventEmitter) EmitEventGeneric(obj object, eventReason, msg string, err error)         {}
func (m *mockEventEmitter) EmitEventRollingDeployWait(obj, k8sObj object, nodeSpecUniqueStr string) {}
func (m *mockEventEmitter) EmitEventOnGetError(obj, getObj object, err error)                       {}
func (m *mockEventEmitter) EmitEventOnUpdate(obj, updateObj object, err error)                      {}
func (m *mockEventEmitter) EmitEventOnDelete(obj, deleteObj object, err error)                      {}
func (m *mockEventEmitter) EmitEventOnCreate(obj, createObj object, err error)                      {}
func (m *mockEventEmitter) EmitEventOnPatch(obj, patchObj object, err error)                        {}
func (m *mockEventEmitter) EmitEventOnList(obj object, listObj objectList, err error)               {}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

func TestDetectVolumeClaimTemplateChanges(t *testing.T) {
	tests := []struct {
		name               string
		currentStatefulSet *appsv1.StatefulSet
		nodeSpec           *v1alpha1.DruidNodeSpec
		expectChanges      bool
		expectDetails      string
	}{
		{
			name: "No changes when annotations are identical",
			currentStatefulSet: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					VolumeClaimTemplates: []v1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "segment-cache",
								Annotations: map[string]string{
									"volume.beta.kubernetes.io/storage-class": "fast-ssd",
								},
							},
						},
					},
				},
			},
			nodeSpec: &v1alpha1.DruidNodeSpec{
				VolumeClaimTemplates: []v1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "segment-cache",
							Annotations: map[string]string{
								"volume.beta.kubernetes.io/storage-class": "fast-ssd",
							},
						},
					},
				},
			},
			expectChanges: false,
			expectDetails: "",
		},
		{
			name: "Detect annotation changes",
			currentStatefulSet: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					VolumeClaimTemplates: []v1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "segment-cache",
								Annotations: map[string]string{
									"volume.beta.kubernetes.io/storage-class": "fast-ssd",
								},
							},
						},
					},
				},
			},
			nodeSpec: &v1alpha1.DruidNodeSpec{
				VolumeClaimTemplates: []v1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "segment-cache",
							Annotations: map[string]string{
								"volume.beta.kubernetes.io/storage-class": "ultra-fast-ssd",
								"backup.policy": "enabled",
							},
						},
					},
				},
			},
			expectChanges: true,
			expectDetails: "VCT segment-cache: annotations changed",
		},
		{
			name: "No changes when VCT is new",
			currentStatefulSet: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					VolumeClaimTemplates: []v1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "segment-cache",
							},
						},
					},
				},
			},
			nodeSpec: &v1alpha1.DruidNodeSpec{
				VolumeClaimTemplates: []v1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "segment-cache",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "new-volume",
							Annotations: map[string]string{
								"new": "annotation",
							},
						},
					},
				},
			},
			expectChanges: false,
			expectDetails: "",
		},
		{
			name: "Detect volumeAttributeClassName changes",
			currentStatefulSet: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					VolumeClaimTemplates: []v1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "segment-cache",
							},
							Spec: v1.PersistentVolumeClaimSpec{
								VolumeAttributesClassName: stringPtr("old-class"),
							},
						},
					},
				},
			},
			nodeSpec: &v1alpha1.DruidNodeSpec{
				VolumeClaimTemplates: []v1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "segment-cache",
						},
						Spec: v1.PersistentVolumeClaimSpec{
							VolumeAttributesClassName: stringPtr("new-class"),
						},
					},
				},
			},
			expectChanges: true,
			expectDetails: "VCT segment-cache: volumeAttributeClassName changed from 'old-class' to 'new-class'",
		},
		{
			name: "Detect both annotation and volumeAttributeClassName changes",
			currentStatefulSet: &appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					VolumeClaimTemplates: []v1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "segment-cache",
								Annotations: map[string]string{
									"backup": "disabled",
								},
							},
							Spec: v1.PersistentVolumeClaimSpec{
								VolumeAttributesClassName: stringPtr("standard"),
							},
						},
					},
				},
			},
			nodeSpec: &v1alpha1.DruidNodeSpec{
				VolumeClaimTemplates: []v1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "segment-cache",
							Annotations: map[string]string{
								"backup": "enabled",
							},
						},
						Spec: v1.PersistentVolumeClaimSpec{
							VolumeAttributesClassName: stringPtr("premium"),
						},
					},
				},
			},
			expectChanges: true,
			expectDetails: "VCT segment-cache: annotations changed; VCT segment-cache: volumeAttributeClassName changed from 'standard' to 'premium'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := detectVolumeClaimTemplateChanges(tt.currentStatefulSet, tt.nodeSpec)
			assert.Equal(t, tt.expectChanges, changes.HasChanges)
			assert.Equal(t, tt.expectDetails, changes.Details)
		})
	}
}

func TestExtractVCTNameFromPVC(t *testing.T) {
	tests := []struct {
		name        string
		pvcName     string
		stsName     string
		expectedVCT string
	}{
		{
			name:        "Extract VCT name from standard PVC",
			pvcName:     "segment-cache-druid-historicals-0",
			stsName:     "druid-historicals",
			expectedVCT: "segment-cache",
		},
		{
			name:        "Extract VCT name with complex naming",
			pvcName:     "data-volume-druid-middlemanager-2",
			stsName:     "druid-middlemanager",
			expectedVCT: "data-volume",
		},
		{
			name:        "Return empty when pattern doesn't match",
			pvcName:     "random-pvc-name",
			stsName:     "druid-historicals",
			expectedVCT: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVCTNameFromPVC(tt.pvcName, tt.stsName)
			assert.Equal(t, tt.expectedVCT, result)
		})
	}
}

// TestPatchStatefulSetVolumeClaimTemplateAnnotations_Integration would be better suited as an integration test
// The actual function requires complex mocking of readers/writers which are global variables

func TestPatchStatefulSetVolumeClaimTemplateAnnotationsDisabled(t *testing.T) {
	ctx := context.TODO()

	druid := &v1alpha1.Druid{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-druid",
			Namespace: "default",
		},
		Spec: v1alpha1.DruidSpec{
			DisablePVCUpdates: true, // Feature disabled
		},
	}

	nodeSpec := &v1alpha1.DruidNodeSpec{
		Kind: "", // Default is StatefulSet
	}

	// Mock event emitter - we just need a no-op implementation
	emitter := &mockEventEmitter{}

	// Test should return immediately when feature is disabled
	deleted, err := patchStatefulSetVolumeClaimTemplates(ctx, nil, druid, nodeSpec, emitter, "test")
	assert.NoError(t, err)
	assert.False(t, deleted) // Should not delete when disabled
}
