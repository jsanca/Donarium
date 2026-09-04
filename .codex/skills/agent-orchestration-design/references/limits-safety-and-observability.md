# Limits, Safety, and Observability

## Control Limits (Defaults to Tune)

| Limit | Typical starting point |
| --- | --- |
| Max agent steps | 10–15 per user request |
| Max tool calls | 20 per request |
| Wall-clock timeout | 60–120s sync; longer async job |
| Token budget per run | Set per product tier |
| Concurrent tool calls | 1–3 unless read-only |

Stop with user-visible message when limits hit — not silent truncation.

## Human-in-the-Loop (HITL)

Require approval before:

- Delete, charge, send external email, publish
- Privilege elevation or cross-tenant access
- Low confidence (classifier score below threshold)

Pattern:

```text
Agent proposes action -> PendingApproval record -> User confirms -> Tool executes with approval_id
```

## Failure Handling

| Failure | Behavior |
| --- | --- |
| Tool timeout | Retry once with backoff; then partial result + error code |
| Tool authz fail | Stop run; do not retry with different args blindly |
| LLM refusal | Return safe message; log for eval |
| Step limit exceeded | Summarize progress; offer continue or human |
| Partial workflow | Persist completed steps; resume from checkpoint |

Design **compensation** only where business requires (saga) — most read-heavy agents need idempotent retries only.

## Observability

Log per run:

- `run_id`, `user_id`, `tenant_id` (tokenized if PII policy requires)
- Each step: model, tokens, latency, tool name, success/fail
- Final outcome and limit-hit reason

Enable support to **replay** a failed run from stored tool outputs (not re-execute mutating tools blindly).

## Runaway Loop Signals

- Same tool called repeatedly with identical args
- Monotonic token growth without user message change
- Oscillation between two tools

Mitigate: dedupe tool calls, circuit breaker, max duplicate detection.

## Eval for Multi-Step Agents

Golden paths should specify:

- Expected tool sequence (order matters)
- Branch taken for conditional flows
- Final structured output

Single-turn Q&A eval is insufficient for orchestrated features.

## Skill Chain

1. `llm-application-architecture` — decide if agent is appropriate
2. `agent-orchestration-design` — topology and state (this skill)
3. `tool-calling-design-review` — each tool contract
4. `ai-evaluation-architecture` — multi-step golden paths
