---
name: tool-calling-design-review
description: Review agent tool and function-calling designs for schema quality, safety, authorization, idempotency, observability, and reliability. Use when designing MCP tools, LLM function APIs, agent actions, or reviewing tool implementations that mutate state or access private data.
---

# Tool Calling Design Review

## Workflow

1. Inventory tools exposed to the model: read-only vs mutating, data scope, and caller identity.
2. Review each tool schema: name, description, parameters, required fields, enums, and examples — the model chooses tools from this metadata.
3. Evaluate safety: authz checks, input validation, side-effect boundaries, and human confirmation for destructive actions.
4. Assess reliability: idempotency keys, timeouts, retries, partial failure handling, and deterministic error shapes.
5. Review observability: tool call logs, latency, success rate, and correlation with conversation/session IDs.
6. Check abuse cases: prompt injection triggering privileged tools, parameter tampering, and excessive call loops.
7. Recommend schema or implementation changes with smallest risk reduction first.

## References

- Read `references/schema-and-contracts.md` for tool naming, JSON schema, and error contracts.
- Read `references/safety-and-reliability.md` for authz, idempotency, confirmation gates, and loop limits.

## Review Checklist

- Tool descriptions state when **not** to use the tool.
- Mutating tools require explicit authorization server-side.
- Parameters are validated; no raw SQL/shell from model input without sandboxing.
- Destructive actions require confirmation or elevated role.
- Idempotency for create/charge/send operations where duplicates hurt.
- Timeouts and rate limits prevent runaway agent loops.
- Tool results are structured and sized — not unbounded dumps into context.
- Tests cover happy path, auth failure, validation failure, and timeout.

## Output

Produce a tool design review with:

- **Tool inventory** — read vs write, owner, dependencies
- **Findings** per tool with exploit or failure scenario
- **Schema recommendations** — description/parameter fixes
- **Safety controls** — auth, confirmation, rate limits
- **Verification** — tests and monitoring to add
