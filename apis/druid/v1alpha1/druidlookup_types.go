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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DruidLookupSpec defines the desired state of DruidLookup
type DruidLookupSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// +required
	DruidClusterName string `json:"druidCluster"`

	// +required
	Id string `json:"id"`

	// +optional
	// +kubebuilder:default:=__default
	Tier string `json:"tier"`

	// +required
	Spec string `json:"spec"`
}

// DruidLookupStatus defines the observed state of DruidLookup
type DruidLookupStatus struct {
	// +optional
	Loaded *bool `json:"loaded,omitempty"`

	// +optional
	PendingNodes []string `json:"pendingNodes,omitempty"`

	// +optional
	NumberOfPendingNodes *int `json:"numberOfPendingNodes,omitempty"`

	// +optional
	LastAppliedSpec string `json:"lastAppliedSpec,omitempty"`

	// +optional
	LastSuccessfulUpdateAt *metav1.Time `json:"lastSuccessfulUpdateAt,omitempty"`

	// +optional
	LastUpdateAttemptAt *metav1.Time `json:"lastUpdateAttemptAt,omitempty"`

	// +optional
	LastUpdateAttemptSuccessful bool `json:"lastUpdateAttemptSuccessful,omitempty"`

	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
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
