/*
Copyright 2022.

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

type ApplicationDomainSSLSpec struct {
	Enabled bool `json:"enabled"`
}

type ApplicationDomainSpec struct {
	Host string                   `json:"host"`
	Path string                   `json:"path"`
	SSL  ApplicationDomainSSLSpec `json:"ssl"`
}

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	Replicas *int32                  `json:"replicas,omitempty"`
	Image    string                  `json:"image"`
	HttpPort int32                   `json:"httpPort"`
	Domains  []ApplicationDomainSpec `json:"domains"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	Ready           bool   `json:"ready,omitempty"`
	CurrentReplicas string `json:"currentReplicas,omitempty"`
	Endpoints       string `json:"endpoints,omitempty"` //Q: Really?
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.status.currentReplicas`
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Port",type=integer,priority=1,JSONPath=`.spec.httpPort`
// +kubebuilder:printcolumn:name="Endpoints",type=string,JSONPath=`.status.endpoints`
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
