# Agent Platform — Budget-Aware Scheduling

Functional mock of an agent platform on Kubernetes that simulates scheduling with budget awareness and automatic reassignment via pool fallback.

## Architecture

```
                ┌───────────────┐
                │ Task Request  │
                └──────┬────────┘
                       │
                       ▼
                 TaskController
                       │
                       ▼
              Budget Scheduler
                       │
           ┌───────────┼───────────┐
           ▼           ▼           ▼
       Agent A     Agent B     Agent C
       Team Pool   Shared      Global
```

**Flow:**
1. A `Task` CR is created with a required skill and a cost
2. The `TaskController` asks the `Scheduler` to assign an agent
3. The `Scheduler` finds agents with the skill, sorts by pool priority (team → shared → global)
4. If the agent has budget → assigns it and deducts the cost
5. If it has no budget → falls back to the next available pool

## Custom Resources

| CRD | Description |
|-----|-------------|
| **Agent** | Executor agent with pool, skills, and budget reference |
| **Skill** | Executable capability with container image and token cost |
| **Budget** | Budget with limit and usage tracking |
| **Task** | Requested task with required skill, cost, and team |

### Agent Example
```yaml
apiVersion: agents.platform/v1
kind: Agent
metadata:
  name: agent-marketing
spec:
  pool: team
  skills: [summarize, translate]
  budgetRef: marketing-budget
  endpoint: http://agent-marketing:8080
```

### Task Example
```yaml
apiVersion: agents.platform/v1
kind: Task
metadata:
  name: summarize-task
spec:
  skill: summarize
  cost: 10
  team: marketing
```

## Fallback Pools

Three levels with priority order:

1. **team** — same team agents (highest priority)
2. **shared** — shared agents across teams
3. **global** — global agents (lowest priority)

When an agent doesn't have enough budget, the scheduler automatically searches the next pool.

## Stack

- **Cluster:** kind (Kubernetes >= 1.29)
- **Runtime:** Go, controller-runtime, kubebuilder
- **Budget state:** in-memory (mock de Redis)
- **Tools:** kubectl, kustomize, make

## Project Structure

```
agent-platform/
├── api/v1/                    # CRD type definitions
│   ├── agent_types.go
│   ├── skill_types.go
│   ├── budget_types.go
│   └── task_types.go
├── internal/controller/       # Kubernetes controllers
│   ├── agent_controller.go
│   ├── budget_controller.go
│   └── task_controller.go
├── scheduler/                 # Budget-aware scheduler
│   ├── scheduler.go
│   └── scheduler_test.go
├── manifests/                 # Sample CRs
│   ├── agents/
│   ├── budgets/
│   ├── skills/
│   └── tasks/
├── config/                    # Kubebuilder config (CRDs, RBAC, deploy)
├── cmd/main.go                # Entrypoint
├── kind-config.yaml           # Kind cluster config
└── Makefile
```

## Requirements

- Go >= 1.21
- Docker
- kind
- kubectl
- kubebuilder
- kustomize
- make

## Quick Start

### 1. Create cluster

```bash
make kind-create
```

### 2. Install CRDs

```bash
make install
```

### 3. Build and deploy the controller

```bash
make docker-build IMG=agent-platform:dev
kind load docker-image agent-platform:dev --name agent-platform
make deploy IMG=agent-platform:dev
```

### 4. Deploy sample resources

```bash
make deploy-samples
```

### 5. Create a task and see the result

```bash
make test-flow
```

Expected output:

```
NAME             SKILL       PHASE       ASSIGNEDAGENT     COST
summarize-task   summarize   scheduled   agent-marketing   10
```

## Fallback Test

Create multiple tasks to exhaust the team budget and observe the fallback:

```bash
# Create 10 tasks with cost 10 (total = 100 = marketing-budget limit)
for i in $(seq 1 10); do
  kubectl apply -f - <<EOF
apiVersion: agents.platform/v1
kind: Task
metadata:
  name: task-$i
spec:
  skill: summarize
  cost: 10
  team: marketing
EOF
done

# Task 11 should fall back to nlp-agent (shared pool)
kubectl apply -f - <<EOF
apiVersion: agents.platform/v1
kind: Task
metadata:
  name: task-overflow
spec:
  skill: summarize
  cost: 10
  team: marketing
EOF

# Verify
kubectl get tasks task-overflow -o wide
# → ASSIGNEDAGENT: nlp-agent (fallback!)

kubectl get budgets -o wide
# → marketing-budget: USED=100, REMAINING=0
# → shared-budget:    USED=10,  REMAINING=490
```

Controller logs:

```
agent budget exceeded, falling back  {"task":"task-overflow", "from":"marketing", "to":"nlp-agent"}
task scheduled  {"task":"task-overflow", "agent":"nlp-agent", "fallback":true}
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, conventions, and how to submit PRs.

## Unit Tests

```bash
go test ./scheduler/ -v
```

```
=== RUN   TestSchedule_HappyPath
=== RUN   TestSchedule_BudgetExhausted_Fallback
=== RUN   TestSchedule_NoBudgetAvailable
=== RUN   TestSchedule_NoAgentsWithSkill
=== RUN   TestSchedule_PoolOrder_TeamSharedGlobal
=== RUN   TestSchedule_PoolFallbackChain
=== RUN   TestSchedule_BudgetDeduction
PASS
```

## Useful Commands

| Command | Description |
|---------|-------------|
| `make kind-create` | Create kind cluster |
| `make kind-delete` | Delete kind cluster |
| `make install` | Install CRDs |
| `make deploy IMG=agent-platform:dev` | Deploy the controller |
| `make deploy-samples` | Deploy agents, budgets, skills |
| `make test-flow` | Create test task and see result |
| `make docker-build IMG=agent-platform:dev` | Build Docker image |
| `make kind-load` | Load image into kind |
| `kubectl get tasks -o wide` | View tasks with assigned agent |
| `kubectl get budgets -o wide` | View budget status |
| `kubectl get agents -o wide` | View agent status |
| `kubectl logs deploy/agent-platform-controller-manager -n agent-platform-system` | View logs |

## Cleanup

```bash
make undeploy
make kind-delete
```

## License

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

