/*
Copyright 2017 The Kubernetes Authors.
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

// +kubebuilder:object:root=true// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.spec.message`
// +kubebuilder:printcolumn:name="Count",type=integer,JSONPath=`.spec.count`

// PR is the Schema for the prs API
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Foo is a specification for a Foo resource
type PipelineRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineRunSpec   `json:"spec,omitempty"`
	Status PipelineRunStatus `json:"status,omitempty"`
}

// PRSpec defines the desired state of PR
type PipelineRunSpec struct {
	Count   int    `json:"count"`
	Message string `json:"message"`
}

// PRStatus defines the observed state of PR
type PipelineRunStatus struct {
	Count   int    `json:"count"`
	Message string `json:"message"`
}

// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// TR is the Schema for the trs API

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Foo is a specification for a Foo resource
type TaskRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TaskRunSpec   `json:"spec,omitempty"`
	Status TaskRunStatus `json:"status,omitempty"`
}

// TRSpec defines the desired state of TR
type TaskRunSpec struct {
	Count   int    `json:"count"`
	Message string `json:"message"`
}

// TRStatus defines the observed state of TR
type TaskRunStatus struct {
	Count   int    `json:"count"`
	Message string `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FooList is a list of Foo resources
type PipelineRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PipelineRun `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// FooList is a list of Foo resources
type TaskRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TaskRun `json:"items"`
}
