/*
Copyright 2026.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// QdrantClusterSpec defines the desired state of QdrantCluster
type QdrantClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// Size is the number of Qdrant nodes to deploy.
	// +kubebuilder:validation:Minimum=1
	Size int32 `json:"size"`

	// Version is the Qdrant docker image tag (e.g., "latest" or "v1.7.0").
	// +kubebuilder:default:="latest"
	Version string `json:"version,omitempty"`

	// StorageSize defines the size of the Persistent Volume for vector data.
	// +kubebuilder:default:="10Gi"
	StorageSize string `json:"storageSize,omitempty"`

	// HTTPPort is the REST API port.
	// +kubebuilder:default:=6333
	HTTPPort int32 `json:"httpPort,omitempty"`

	// GRPCPort is the gRPC API port.
	// +kubebuilder:default:=6334
	GRPCPort int32 `json:"grpcPort,omitempty"`
}

// QdrantClusterStatus defines the observed state of QdrantCluster.
type QdrantClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// ReadyReplicas shows how many Qdrant nodes are currently running and ready.
	ReadyReplicas int32 `json:"readyReplicas"`

	// ClusterState indicates the current health (e.g., "Provisioning", "Ready", "Failed")
	ClusterState string `json:"clusterState,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// QdrantCluster is the Schema for the qdrantclusters API
type QdrantCluster struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of QdrantCluster
	// +required
	Spec QdrantClusterSpec `json:"spec"`

	// status defines the observed state of QdrantCluster
	// +optional
	Status QdrantClusterStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// QdrantClusterList contains a list of QdrantCluster
type QdrantClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []QdrantCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QdrantCluster{}, &QdrantClusterList{})
}
