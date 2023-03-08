package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type DruidIngestionMethod string

const (
	Kafka                    DruidIngestionMethod = "kafka"
	Kinesis                  DruidIngestionMethod = "kinesis"
	NativeBatchIndexParallel DruidIngestionMethod = "native-batch"
	QueryControllerSQL       DruidIngestionMethod = "sql"
	HadoopIndexHadoop        DruidIngestionMethod = "index-hadoop"
)

type DruidIngestionSpec struct {
	RouterURL     string                     `json:"routerUrl"`
	Suspend       bool                       `json:"suspend"`
	IngestionSpec DruidOperatorIngestionSpec `json:"ingestionSpec"`
}

type DruidOperatorIngestionSpec struct {
	Type           DruidIngestionMethod `json:"type"`
	SupervisorSpec string               `json:"supervisorSpec,omitempty"`
}

type DruidIngestionStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.ingestionSpec.type"
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
