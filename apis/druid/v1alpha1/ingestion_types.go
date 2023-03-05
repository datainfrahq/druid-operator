package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type DruidIngestionSpec struct {
	RouterURL      string `json:"routerUrl"`
	Suspend        bool   `json:"suspend"`
	SupervisorSpec string `json:"supervisorSpec"`
}

type DruidIngestionStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// Ingestion is the Schema for the Ingestion API
type DruidIngestion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DruidIngestionSpec   `json:"spec"`
	Status DruidIngestionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// IngestionList contains a list of Ingestion
type DruidIngestionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DruidIngestion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DruidIngestion{}, &DruidIngestionList{})
}
