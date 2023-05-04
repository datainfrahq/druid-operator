/*
DataInfra 2023 Copyright 2023.

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
)

type DruidIngestionMethod string

const (
	Kafka                    DruidIngestionMethod = "kafka"
	Kinesis                  DruidIngestionMethod = "kinesis"
	NativeBatchIndexParallel DruidIngestionMethod = "native-batch"
	QueryControllerSQL       DruidIngestionMethod = "sql"
	HadoopIndexHadoop        DruidIngestionMethod = "index-hadoop"
)

type DruidIngestionSpec struct {
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
