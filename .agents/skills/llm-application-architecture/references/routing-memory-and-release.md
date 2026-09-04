# Model Routing, Memory, and Phased Release

## Model Routing

| Tier | Use | Example |
| --- | --- | --- |
| Fast / cheap | Routing, classification, extraction | Haiku, mini models |
| Standard | Most user-facing chat | Sonnet-class |
| Strong | Complex reasoning, low volume | Opus-class |

**Fallback chain:** primary -> secondary provider/model -> cached/template response -> graceful error.

Log route reason (intent, token estimate, failure) for cost tuning.

## Context Budget (Drop Order)

When approaching context limits, drop in this order unless product says otherwise:

1. Oldest chat turns (keep summary if needed)
2. Redundant tool payloads (summarize on server)
3. Lower-ranked retrieval chunks
4. Verbose reference docs (keep citations metadata)
5. Never drop: system policy, auth context, current user task

## Memory Layers

| Layer | Scope | Storage | Risk |
| --- | --- | --- | --- |
| Thread | Single conversation | Session/DB | Low if scoped |
| User profile | Preferences, locale | User store | PII — minimize |
| Corpus | Knowledge | Vector/keyword index | ACL at retrieval |
| Agent state | Workflow step | Workflow engine | Must be durable for retries |

Do not use long thread history as a substitute for RAG or tools.

## Sync vs Async

| Mode | When | UX |
| --- | --- | --- |
| Sync streaming | Chat, copilot panels | SSE/WebSocket tokens |
| Sync blocking | Short classify/extract | Spinner <5s |
| Async job | Reports, bulk summarization | Job ID + status |
| Human-in-the-loop | Destructive tools, low confidence | Approval queue |

## Phased Rollout Template

**Phase 1 — Prove value**

- One pattern (usually RAG *or* read-only tools)
- Single model + fallback message
- Manual eval set; no auto-deploy gate yet

**Phase 2 — Harden**

- Eval in CI or pre-release; retrieval ACL tests
- Observability: latency, tokens, retrieval hit rate
- Rate limits and cost caps

**Phase 3 — Scale**

- Hybrid retrieval, rerank, router
- Multi-tenant isolation audits
- Optional orchestration for multi-step flows

## Architecture Brief Snippet (Copy-Ready)

```markdown
## LLM feature: [name]
- Job: [user outcome]
- Pattern: [RAG | tools | hybrid | prompt-only]
- Models: primary [X], fallback [Y]
- Context: system + [retrieval k] + [tool budget]
- SLA: p95 latency [N]s, max cost/request [$]
- MVP scope: [in] / Out of scope: [out]
- Next review: [rag-architecture-review | tool-calling-design-review | ai-evaluation-architecture]
```
