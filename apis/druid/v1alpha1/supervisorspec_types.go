/*

 */

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SupervisorSpecSpec defines the desired state of SupervisorSpec
type SupervisorSpecSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SupervisorSpec. Edit supervisorspec_types.go to remove/update
	ClusterRef     string `json:"clusterRef,omitempty"`
	Suspend        bool   `json:"suspend,omitempty"`
	SupervisorSpec string `json:"supervisorSpec,omitempty"`
}

// SupervisorSpecStatus defines the observed state of SupervisorSpec
type SupervisorSpecStatus struct {
	// Whether or not the CR is in sync with druids api representation
	Synced string `json:"synced"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="InSync",type=string,JSONPath=`.status.synced`

// SupervisorSpec is the Schema for the supervisorspecs API
type SupervisorSpec struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SupervisorSpecSpec   `json:"spec,omitempty"`
	Status SupervisorSpecStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SupervisorSpecList contains a list of SupervisorSpec
type SupervisorSpecList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SupervisorSpec `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SupervisorSpec{}, &SupervisorSpecList{})
}
