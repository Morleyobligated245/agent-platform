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

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentsv1 "github.com/ezequiel/agent-platform/api/v1"
	"github.com/ezequiel/agent-platform/scheduler"
)

// AgentReconciler reconciles an Agent object.
type AgentReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Scheduler *scheduler.Scheduler
}

// +kubebuilder:rbac:groups=agents.platform,resources=agents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=agents.platform,resources=agents/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=agents.platform,resources=agents/finalizers,verbs=update
// +kubebuilder:rbac:groups=agents.platform,resources=budgets,verbs=get;list;watch

// Reconcile handles Agent CR events.
func (r *AgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	var agent agentsv1.Agent
	if err := r.Get(ctx, req.NamespacedName, &agent); err != nil {
		if errors.IsNotFound(err) {
			log.Info("agent deleted, unregistering from scheduler", "agent", req.Name)
			r.Scheduler.UnregisterAgent(req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Register or update the agent in the scheduler.
	r.Scheduler.RegisterAgent(agent)
	log.Info("agent registered in scheduler", "agent", agent.Name, "pool", agent.Spec.Pool)

	// Determine agent state based on budget remaining.
	state := "ready"
	if agent.Spec.BudgetRef != "" {
		var budget agentsv1.Budget
		budgetKey := types.NamespacedName{
			Name:      agent.Spec.BudgetRef,
			Namespace: agent.Namespace,
		}
		if err := r.Get(ctx, budgetKey, &budget); err != nil {
			if !errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			// Budget not found; keep state ready.
		} else {
			remaining := budget.Spec.Limit - budget.Status.Used
			if remaining <= 0 {
				state = "exhausted"
			}
		}
	}

	agent.Status.State = state
	if err := r.Status().Update(ctx, &agent); err != nil {
		return ctrl.Result{}, err
	}

	log.Info("agent status updated", "agent", agent.Name, "state", state)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentsv1.Agent{}).
		Named("agent").
		Complete(r)
}
