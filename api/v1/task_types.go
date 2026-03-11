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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TaskSpec defines the desired state of Task
type TaskSpec struct {
	// skill is the name of the Skill required to execute this task.
	// +required
	Skill string `json:"skill"`

	// cost is the token cost of executing this task.
	// +kubebuilder:validation:Minimum=1
	// +required
	Cost int `json:"cost"`

	// team is the team that owns this task.
	// +required
	Team string `json:"team"`
}

// TaskStatus defines the observed state of Task.
type TaskStatus struct {
	// phase indicates the current lifecycle phase of the task.
	// +kubebuilder:validation:Enum=pending;scheduled;completed;failed
	// +optional
	Phase string `json:"phase,omitempty"`

	// assignedAgent is the name of the Agent assigned to execute this task.
	// +optional
	AssignedAgent string `json:"assignedAgent,omitempty"`

	// reason provides a human-readable explanation for the current phase.
	// +optional
	Reason string `json:"reason,omitempty"`

	// conditions represent the current state of the Task resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Skill",type=string,JSONPath=`.spec.skill`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="AssignedAgent",type=string,JSONPath=`.status.assignedAgent`
// +kubebuilder:printcolumn:name="Cost",type=integer,JSONPath=`.spec.cost`

// Task is the Schema for the tasks API
type Task struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Task
	// +required
	Spec TaskSpec `json:"spec"`

	// status defines the observed state of Task
	// +optional
	Status TaskStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// TaskList contains a list of Task
type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Task `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Task{}, &TaskList{})
}
