---
name: agent-orchestration-design
description: Design multi-step agent and workflow architecture. Use when planning agent loops, subagents, supervisor patterns, handoffs, durable state, human approval, or deciding whether a deterministic pipeline is enough instead of an autonomous agent.
---

# Agent Orchestration Design

## Workflow

1. State the **user outcome** and whether it requires branching, external systems, or fixed steps — agents excel at branching; pipelines excel at fixed ETL-style flows.
2. Apply the **complexity gate**: if steps are known and linear, prefer a workflow engine or plain code over an LLM agent loop.
3. Choose topology: single agent with tools, router + specialists, supervisor + workers, or handoff between role agents.
4. Define **state ownership**: conversation transcript, workflow instance, tool results, and idempotency keys — what persists across retries.
5. Set **control limits**: max steps, max tool calls, timeouts, cost cap, and when to escalate to human.
6. Plan failure handling: retry which steps, compensate partial work, surface partial results to user.
7. Review observability: trace per step, tool call graph, and debug replay for failed runs.
8. Cross-check tool design (`tool-calling-design-review`) and eval (`ai-evaluation-architecture`) for multi-step paths.

## References

- Read `references/topologies-and-state.md` for single vs router vs supervisor patterns and state stores.
- Read `references/limits-safety-and-observability.md` for step caps, HITL, failure modes, and tracing.

## Design Checklist

- A deterministic workflow was considered and rejected with reason — or chosen instead.
- Each agent role has a narrow goal and bounded tools.
- State is durable where retries or long runs matter.
- Mutating tools require authz and idempotency (see tool-calling skill).
- Max steps and cost limits prevent runaway loops.
- Human approval defined for destructive or high-risk branches.
- User sees progress (status) for multi-step runs — not a silent hang.
- Eval covers multi-step golden paths — not only single-turn cases.

## Output

Deliver an agent orchestration design with:

- **Topology diagram** — agents, routers, tools, data stores
- **Step/state model** — what is stored where between steps
- **Control policy** — limits, timeouts, escalation
- **Failure and retry** — per step behavior
- **Simpler alternative** — pipeline or single-call option if applicable
- **Implementation notes** — sync vs async job, idempotency
- **Follow-on reviews** — tools, eval, security as needed
