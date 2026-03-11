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

// AgentSpec defines the desired state of Agent
type AgentSpec struct {
	// pool indicates the agent pool (team, shared, or global).
	// +kubebuilder:validation:Enum=team;shared;global
	// +required
	Pool string `json:"pool"`

	// skills is the list of Skill names this agent can execute.
	// +required
	Skills []string `json:"skills"`

	// budgetRef is the name of the Budget resource this agent draws from.
	// +required
	BudgetRef string `json:"budgetRef"`

	// endpoint is an optional URL where the agent can be reached.
	// +optional
	Endpoint string `json:"endpoint,omitempty"`
}

// AgentStatus defines the observed state of Agent.
type AgentStatus struct {
	// state reflects the current operational state of the agent.
	// +kubebuilder:validation:Enum=ready;busy;exhausted
	// +optional
	State string `json:"state,omitempty"`

	// assignedTasks is the number of tasks currently assigned to this agent.
	// +kubebuilder:default=0
	// +optional
	AssignedTasks int `json:"assignedTasks,omitempty"`

	// conditions represent the current state of the Agent resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Pool",type=string,JSONPath=`.spec.pool`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="BudgetRef",type=string,JSONPath=`.spec.budgetRef`

// Agent is the Schema for the agents API
type Agent struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Agent
	// +required
	Spec AgentSpec `json:"spec"`

	// status defines the observed state of Agent
	// +optional
	Status AgentStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// AgentList contains a list of Agent
type AgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Agent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Agent{}, &AgentList{})
}
