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

// SkillCost defines the cost associated with executing a skill.
type SkillCost struct {
	// cpu is the CPU cost for executing this skill.
	// +optional
	CPU string `json:"cpu,omitempty"`

	// tokens is the token cost for executing this skill.
	// +required
	Tokens int `json:"tokens"`
}

// SkillSpec defines the desired state of Skill
type SkillSpec struct {
	// containerImage is the OCI image that implements this skill.
	// +required
	ContainerImage string `json:"containerImage"`

	// cost defines the resource cost of executing this skill.
	// +required
	Cost SkillCost `json:"cost"`
}

// SkillStatus defines the observed state of Skill.
type SkillStatus struct {
	// conditions represent the current state of the Skill resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ContainerImage",type=string,JSONPath=`.spec.containerImage`
// +kubebuilder:printcolumn:name="Tokens",type=integer,JSONPath=`.spec.cost.tokens`

// Skill is the Schema for the skills API
type Skill struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Skill
	// +required
	Spec SkillSpec `json:"spec"`

	// status defines the observed state of Skill
	// +optional
	Status SkillStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// SkillList contains a list of Skill
type SkillList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Skill `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Skill{}, &SkillList{})
}
