# LLM Application Pattern Selection

## Quick Decision Tree

```text
Need factual answers from private/changing docs?
  yes -> RAG (maybe + tools for actions)
  no  -> Need to call APIs/DB/write state?
          yes -> Tools / agent with tools
          no  -> Prompt-only (+ optional cache) sufficient?
                  yes -> Single-shot or short thread LLM call
                  no  -> Revisit requirements (often hidden RAG/tools need)
```

Fine-tuning is rarely step one. Prefer RAG/tools/prompt engineering until eval proves a format/style gap that data can fix.

## Pattern Comparison

| Pattern | Best for | Weak when | Ops burden |
| --- | --- | --- | --- |
| Prompt-only | Summarize, classify, rewrite public text | Private facts, live data | Low |
| RAG | Q&A over docs, catalogs, policies | Actions, real-time transactional state | Medium–high |
| Tools / functions | CRUD, search APIs, workflows | Large corpus Q&A without retrieval | Medium |
| Agent loop | Multi-step tasks with branching | Simple single-call tasks | High |
| Fine-tuning | Stable style/format, domain jargon | Facts that change weekly | High |

## Hybrid Layouts (Common in Production)

### RAG + tools

Retrieve context for grounding; tools for live lookups and mutations.

```text
User -> API -> retrieve(top_k) -> LLM with tools -> tool calls (authorized) -> final answer + citations
```

### Router + specialists

Cheap model classifies intent; routes to RAG path, tool path, or refusal.

```text
User -> router (small model or rules) -> {rag_handler | tool_handler | support_handoff}
```

### Async enrichment

User gets fast acknowledgment; job runs retrieval + generation; result pushed via poll/WebSocket/email.

Use when latency SLO exceeds model+retrieval time (e.g. >30s reports).

## Anti-Patterns

| Anti-pattern | Why it fails |
| --- | --- |
| "RAG for everything" | Wrong for transactional updates; stale corpus |
| "Agent for a single API call" | Extra failure modes without benefit |
| "Bigger context window instead of retrieval" | Cost, latency, lost-in-middle |
| "Fine-tune first" | Expensive; facts drift; hard to audit |
| "Model remembers tenant rules" | Must enforce in retrieval/tools server-side |

## Minimal MVP by Feature Type

| Feature | MVP | Defer |
| --- | --- | --- |
| Docs Q&A | RAG + citations + 20–50 golden questions | Rerank, hybrid, multi-index |
| Copilot in app | Tools for read-only + prompt with UI context | Write tools until authz solid |
| Support triage | Classify + route; optional RAG for macros | Full autonomous resolution |
| Code assistant | IDE context + tools (repo scoped) | Custom fine-tunes |
