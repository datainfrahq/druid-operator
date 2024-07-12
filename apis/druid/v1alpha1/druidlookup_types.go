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

package v1alpha1

import (
	"encoding/json"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DruidLookupSpec defines the desired state of DruidLookup
type DruidLookupSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// The name of the druid cluster to apply the lookup in.
	//
	// Assumed to be within the same k8s namespace.
	// +required
	DruidCluster v1.LocalObjectReference `json:"druidCluster"`

	// The tier to put the lookup in.
	// +optional
	// +kubebuilder:default:=__default
	Tier string `json:"tier"`

	// Lookup template.
	//
	// Any stringified json value that is applicable in the `lookupExtractorFactory` field.
	//
	// Please see https://druid.apache.org/docs/latest/api-reference/lookups-api#update-lookup.
	// +required
	Template string `json:"template"`
}

// DruidLookupStatus defines the observed state of DruidLookup
type DruidLookupStatus struct {

	// `true` if the druid cluster has reported that the lookup is loaded on all relevant nodes, otherwise false.
	// +optional
	Loaded *bool `json:"loaded,omitempty"`

	// A list of the nodes that the druid cluster reports to yet have loaded the lookup.
	// +optional
	PendingNodes []string `json:"pendingNodes,omitempty"`

	// The number of nodes that the druid cluster reports to yet have loaded the lookup.
	//
	// (Exists in conjunction with `PendingNodes` to facilitate displaying this summary using kubebuilders print column feature.)
	// +optional
	NumberOfPendingNodes *int `json:"numberOfPendingNodes,omitempty"`

	// The druid cluster that the last successful application of this lookup happened in.
	//
	// Used to determine if changes require old lookup to be deleted.
	// +optional
	LastClusterAppliedIn v1.LocalObjectReference `json:"lastClusterAppliedIn"`

	// The tier that the last successful application of this lookup happened in.
	//
	// Used to determine if changes require old lookup to be deleted.
	// +optional
	LastTierAppliedIn string `json:"lastTierAppliedIn"`

	// The template that the last successful application of this lookup applied.
	//
	// Used to determine if changes require old lookup to be deleted.
	// +optional
	LastAppliedTemplate string `json:"lastAppliedTemplate,omitempty"`

	// The time that the last successful application of this lookup happened at.
	// +optional
	LastSuccessfulUpdateAt *metav1.Time `json:"lastSuccessfulUpdateAt,omitempty"`

	// The time that the last attempt to apply this lookup happened at.
	// +optional
	LastUpdateAttemptAt *metav1.Time `json:"lastUpdateAttemptAt,omitempty"`

	// `true` if the last application attempt were successful, `false` otherwise.
	// +optional
	LastUpdateAttemptSuccessful bool `json:"lastUpdateAttemptSuccessful,omitempty"`

	// If the last application attempt failed, this property contains the associated error message, otherwise unset.
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Target cluster",JSONPath=`.spec.druidCluster.name`,type=string
//+kubebuilder:printcolumn:name="Target tier",JSONPath=`.spec.tier`,type=string
//+kubebuilder:printcolumn:name="Ok",JSONPath=`.status.lastUpdateAttemptSuccessful`,type=boolean
//+kubebuilder:printcolumn:name="Loaded",JSONPath=`.status.loaded`,type=boolean,priority=1
//+kubebuilder:printcolumn:name="Pending Nodes",JSONPath=`.status.numberOfPendingNodes`,type=integer,priority=10
//+kubebuilder:printcolumn:name="Updated At",JSONPath=`.status.lastSuccessfulUpdateAt`,type=date,priority=5
//+kubebuilder:printcolumn:name="Age",JSONPath=`.metadata.creationTimestamp`,type=date

// DruidLookup is the Schema for the druidlookups API
type DruidLookup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DruidLookupSpec   `json:"spec,omitempty"`
	Status DruidLookupStatus `json:"status,omitempty"`
}

func (dl *DruidLookup) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: dl.Namespace,
		Name:      dl.Name,
	}
}

func (dl *DruidLookup) ShouldDeleteLastAppliedLookup() bool {
	hasAppliedBefore := dl.Status.LastClusterAppliedIn.Name != "" && dl.Status.LastTierAppliedIn != ""
	if !hasAppliedBefore {
		return false
	}

	if dl.Status.LastClusterAppliedIn.Name != dl.Spec.DruidCluster.Name {
		return true
	}

	if dl.Status.LastTierAppliedIn != dl.Spec.Tier {
		return true
	}

	return false
}

func (dl *DruidLookup) GetTemplateToApply() (interface{}, error) {
	var currentTemplate interface{}
	if err := json.Unmarshal([]byte(dl.Spec.Template), &currentTemplate); err != nil {
		return nil, err
	}

	if dl.Status.LastAppliedTemplate == "" {
		return currentTemplate, nil
	}

	clusterChanged := dl.Spec.DruidCluster.Name != dl.Status.LastClusterAppliedIn.Name
	tierChanged := dl.Spec.Tier != dl.Status.LastTierAppliedIn

	if clusterChanged || tierChanged {
		return currentTemplate, nil
	}

	var oldTemplate interface{}
	if err := json.Unmarshal([]byte(dl.Status.LastAppliedTemplate), &oldTemplate); err != nil {
		return nil, err
	}

	if !reflect.DeepEqual(currentTemplate, oldTemplate) {
		return currentTemplate, nil
	}

	return nil, nil
}

//+kubebuilder:object:root=true

// DruidLookupList contains a list of DruidLookup
type DruidLookupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DruidLookup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DruidLookup{}, &DruidLookupList{})
}
