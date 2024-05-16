/*
Copyright 2024 parvejmia9.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BookstoreStatus defines the observed state of Bookstore
type BookstoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Bookstore is the Schema for the bookstores API
type Bookstore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BookStoreSpec   `json:"spec,omitempty"`
	Status BookstoreStatus `json:"status,omitempty"`
}

type BookStoreStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

//+kubebuilder:object:root=true

// BookstoreList contains a list of Bookstore
type BookstoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bookstore `json:"items"`
}

// BookStoreSpec defines the desired state of Bookstore
type BookStoreSpec struct {
	DeploymentSpec DeploymentSpec `json:"deployment-spec,omitempty"`
	ServiceSpec    ServiceSpec    `json:"service-spec,omitempty"`
}

// DeploymentSpec contains specs of container
type DeploymentSpec struct {
	Name     string `json:"name,omitempty"`
	Replicas *int32 `json:"replicas"`
	Image    string `json:"image,omitempty"`
	Port     int32  `json:"port,omitempty"`
}

type ServiceSpec struct {
	// +optional
	Name string `json:"name,omitempty"`
	// +optional
	Type corev1.ServiceType `json:"type,omitempty"`
	Port int32              `json:"port,omitempty"`
	// +optional
	TargetPort int32 `json:"targetPort,omitempty"`
	// +optional
	NodePort int32 `json:"nodePort,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Bookstore{}, &BookstoreList{})
}
