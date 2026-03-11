# Architecture

This document describes the internal architecture, design decisions, and data flow of the Agent Platform.

## Overview

The Agent Platform is a Kubernetes operator that implements budget-aware scheduling for AI agents. It uses Custom Resource Definitions (CRDs) to model agents, skills, budgets, and tasks, with a central scheduler that handles assignment and fallback logic.

## System Design

```
                    ┌─────────────────────┐
                    │   Kubernetes API    │
                    └──────────┬──────────┘
                               │ watch/update
              ┌────────────────┼────────────────┐
              │                │                │
     ┌────────▼───────┐ ┌─────▼──────┐ ┌───────▼──────┐
     │ AgentController│ │BudgetCtrl  │ │ TaskController│
     └────────┬───────┘ └─────┬──────┘ └───────┬──────┘
              │               │                 │
              │    ┌──────────▼──────────┐      │
              └───►│     Scheduler       │◄─────┘
                   │  (shared instance)  │
                   │                     │
                   │  ┌──────────────┐   │
                   │  │ Agent Map    │   │
                   │  │ Budget Map   │   │
                   │  └──────────────┘   │
                   └─────────────────────┘
```

## Custom Resources

### Agent

Represents an executor capable of performing skills.

| Field | Type | Description |
|-------|------|-------------|
| `spec.pool` | `team\|shared\|global` | Pool level for fallback ordering |
| `spec.skills` | `[]string` | List of skill names this agent can execute |
| `spec.budgetRef` | `string` | Name of the Budget CR this agent draws from |
| `spec.endpoint` | `string` | Optional URL for the agent service |
| `status.state` | `ready\|busy\|exhausted` | Current state based on budget |
| `status.assignedTasks` | `int` | Number of tasks currently assigned |

### Skill

Describes a capability with associated cost.

| Field | Type | Description |
|-------|------|-------------|
| `spec.containerImage` | `string` | OCI image implementing the skill |
| `spec.cost.cpu` | `string` | CPU cost (e.g., "100m") |
| `spec.cost.tokens` | `int` | Token cost for executing the skill |

### Budget

Controls spending limits per team or pool.

| Field | Type | Description |
|-------|------|-------------|
| `spec.limit` | `int` | Maximum token budget |
| `spec.pool` | `team\|shared\|global` | Which pool this budget applies to |
| `status.used` | `int` | Tokens consumed so far |
| `status.remaining` | `int` | Computed: limit - used |

### Task

Represents a work request to be scheduled.

| Field | Type | Description |
|-------|------|-------------|
| `spec.skill` | `string` | Required skill name |
| `spec.cost` | `int` | Token cost (min 1) |
| `spec.team` | `string` | Team that owns this task |
| `status.phase` | `pending\|scheduled\|completed\|failed` | Lifecycle phase |
| `status.assignedAgent` | `string` | Agent name assigned by scheduler |
| `status.reason` | `string` | Human-readable explanation |

## Scheduling Algorithm

### Pool Priority

Agents are organized into three pools with strict fallback ordering:

| Priority | Pool | Use case |
|----------|------|----------|
| 0 (highest) | `team` | Dedicated team agents |
| 1 | `shared` | Shared across teams |
| 2 (lowest) | `global` | Last resort |

### Scheduling Flow

```
1. Task CR created (phase: "")
2. TaskController picks it up
3. Set phase = "pending"
4. Call scheduler.Schedule(task):
   a. Find all agents with matching skill
   b. Sort by pool priority (team → shared → global)
   c. For each agent in order:
      - Check if budget has enough remaining for task cost
      - If yes: deduct cost, return agent (done)
      - If no: add to exhausted list, continue
   d. If no agent found: return error
5. On success:
   - Set phase = "scheduled", assignedAgent, reason
   - Update Budget CR (used += cost)
6. On failure:
   - Set phase = "failed", reason = error
7. Update Task status subresource
```

### Fallback Example

```
Budget state:
  marketing-budget: limit=100, used=100 (exhausted)
  shared-budget:    limit=500, used=50  (available)

Agents:
  agent-marketing: pool=team,   skills=[summarize], budgetRef=marketing-budget
  nlp-agent:       pool=shared, skills=[summarize], budgetRef=shared-budget

Task: skill=summarize, cost=10

Scheduler:
  1. Find agents with "summarize" → [agent-marketing, nlp-agent]
  2. Sort by pool → [agent-marketing (team=0), nlp-agent (shared=1)]
  3. agent-marketing: budget remaining = 0 < 10 → skip
  4. nlp-agent: budget remaining = 450 >= 10 → assign ✓
  
Result: nlp-agent (fallback=true)
Log: "agent-marketing budget exceeded, fallback to nlp-agent, task assigned"
```

## Controller Design

### AgentController

**Watches:** Agent CRs
**Responsibilities:**
- Register/unregister agents in the shared scheduler
- Fetch referenced Budget CR to determine agent state
- Set `status.state` to "exhausted" when budget ≤ 0, "ready" otherwise

### BudgetController

**Watches:** Budget CRs
**Responsibilities:**
- Sync budget data (limit, used) to the scheduler's in-memory store
- Compute and update `status.remaining = limit - used`

### TaskController

**Watches:** Task CRs
**Responsibilities:**
- Skip tasks already in terminal state (scheduled/completed/failed)
- Call the scheduler to find an appropriate agent
- On success: update task status, deduct cost from Budget CR status
- On failure: mark task as failed with reason

### Shared Scheduler

All three controllers share a single `*scheduler.Scheduler` instance created in `cmd/main.go`. The scheduler maintains:

- **Agent registry** (`map[string]Agent`): populated by AgentController
- **Budget state** (`map[string]int`): populated by BudgetController, consumed by TaskController
- **Thread-safe**: all operations protected by `sync.RWMutex`

## Design Decisions

### Why in-memory instead of Redis?

For the MVP, the scheduler keeps budget state in Go maps. This simplifies deployment (no external dependencies) and is sufficient for validating the scheduling model. The scheduler interface is designed so Redis can be plugged in later without changing controllers.

### Why no NATS?

Same reasoning — the MVP validates CRDs, scheduling, and fallback without needing a task queue. Tasks are processed synchronously via controller-runtime's reconciliation loop, which already provides watch + retry semantics.

### Why a shared scheduler instead of a separate deployment?

Keeps the system simple: one binary, one pod. The scheduler is a Go struct injected into controllers. This avoids network calls and serialization overhead. For production, it could be extracted into a gRPC service.

### Why kubebuilder?

Standard tooling for building Kubernetes operators in Go. Generates CRDs, RBAC, DeepCopy, Dockerfiles, and Makefile automatically. The project follows kubebuilder conventions for easier onboarding.

## Future Architecture (Post-MVP)

### Phase 2: Real Infrastructure
- Replace in-memory budget with **Redis** for persistence
- Add **NATS** for async task dispatch
- Add **Prometheus** metrics for scheduling decisions

### Phase 3: Scale
- Multi-cluster agent federation
- GPU-aware agents
- LLM routing based on model capabilities

### Phase 4: Economics
- Cost-based scheduling optimization
- Internal agent marketplace
- Dynamic budget allocation
