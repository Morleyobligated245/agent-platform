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

// BudgetSpec defines the desired state of Budget
type BudgetSpec struct {
	// limit is the maximum token budget allowed.
	// +kubebuilder:validation:Minimum=0
	// +required
	Limit int `json:"limit"`

	// pool indicates which pool this budget applies to.
	// +kubebuilder:validation:Enum=team;shared;global
	// +required
	Pool string `json:"pool"`
}

// BudgetStatus defines the observed state of Budget.
type BudgetStatus struct {
	// used is the amount of budget consumed so far.
	// +kubebuilder:default=0
	// +optional
	Used int `json:"used,omitempty"`

	// remaining is the amount of budget still available.
	// +optional
	Remaining int `json:"remaining,omitempty"`

	// conditions represent the current state of the Budget resource.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Limit",type=integer,JSONPath=`.spec.limit`
// +kubebuilder:printcolumn:name="Used",type=integer,JSONPath=`.status.used`
// +kubebuilder:printcolumn:name="Remaining",type=integer,JSONPath=`.status.remaining`
// +kubebuilder:printcolumn:name="Pool",type=string,JSONPath=`.spec.pool`

// Budget is the Schema for the budgets API
type Budget struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Budget
	// +required
	Spec BudgetSpec `json:"spec"`

	// status defines the observed state of Budget
	// +optional
	Status BudgetStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// BudgetList contains a list of Budget
type BudgetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Budget `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Budget{}, &BudgetList{})
}
