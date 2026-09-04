# Agent Topologies and State

## When Not to Use an Agent

Prefer **deterministic workflow** (code, Step Functions, Temporal, queue workers) when:

- Steps are fixed and known at design time
- Branching is rule-based (if status=X then Y)
- LLM is only needed for one classify/extract step

Use **agent loop** when:

- Tool choice depends on intermediate results
- User intent requires exploratory search
- Multiple specialized capabilities must coordinate

## Topology Patterns

### Single agent + tools

One LLM loop calls tools until done or limit hit.

```text
User -> Agent (LLM + tool loop) -> Tools -> Final response
```

Best for: copilot with moderate branching, ≤10 steps typical.

### Router + specialists

Router (rules or small model) picks handler; each handler may be RAG-only, tool-only, or mini-agent.

```text
User -> Router -> {SupportRAG | BillingTools | HandoffHuman}
```

Best for: distinct intents with different risk profiles.

### Supervisor + workers

Supervisor delegates subtasks to worker agents with isolated context; supervisor merges results.

```text
User -> Supervisor -> [ResearchWorker | CodeWorker | ReviewWorker] -> Supervisor -> Answer
```

Best for: parallel subtasks; watch token cost and context duplication.

### Handoff agents

Agent A completes phase and passes structured state to Agent B (sales -> implementation).

State must be **schema-defined**, not only raw chat history.

## State Ownership

| State type | Store | Notes |
| --- | --- | --- |
| Chat transcript | Session DB / cache | Trim/summarize for long runs |
| Workflow instance | Workflow engine | step, status, correlation_id |
| Tool results | Append to run log | Summarize before re-prompt |
| User approvals | Queue / ticket | Required for destructive ops |
| Idempotency keys | Per tool call | Survive retries |

Avoid: only storing state in LLM context — retries lose work.

## Handoff Payload Example

```json
{
  "run_id": "run_abc",
  "phase": "research_complete",
  "findings": ["..."],
  "open_questions": ["..."],
  "artifacts": [{"type": "doc_id", "value": "policy-42"}]
}
```

## Comparison Table

| Pattern | Complexity | Debuggability | Cost |
| --- | --- | --- | --- |
| Single + tools | Low | Medium | Medium |
| Router | Medium | High | Low–medium |
| Supervisor | High | Low–medium | High |
| Handoff | Medium | High | Medium |
