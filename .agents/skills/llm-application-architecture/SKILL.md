---
name: llm-application-architecture
description: Choose LLM application patterns before implementation. Use when adding chat, copilot, or Q&A to a product, comparing RAG vs tools vs prompt-only, selecting model routing and fallbacks, or reviewing an AI feature design with developers and architects.
---

# LLM Application Architecture

## Workflow

1. Clarify the user job: question answering, task execution, drafting, search, or multi-step workflow — and success criteria (accuracy, latency, cost, safety).
2. Classify data needs: static knowledge, live systems, user-specific context, or none — and freshness/ACL requirements.
3. Compare patterns: prompt-only, RAG, tool/function calling, fine-tuning, or hybrid — with explicit tradeoffs for this feature.
4. Decide interaction model: sync chat, streaming UI, async jobs, batch, or human-in-the-loop approval gates.
5. Define model strategy: primary/fallback models, caching, routing by task complexity, and token/cost budgets.
6. Sketch context architecture: system prompt, retrieved docs, tool results, memory layers, and what to drop first under pressure.
7. Identify follow-on reviews: RAG pipeline (`rag-architecture-review`), tools (`tool-calling-design-review`), orchestration (`agent-orchestration-design`), eval (`ai-evaluation-architecture`).
8. Recommend a **minimal first release** and what to defer.

## References

- Read `references/pattern-selection.md` for RAG vs tools vs prompt-only decision trees and hybrid layouts.
- Read `references/routing-memory-and-release.md` for model routing, context budgets, memory, and phased rollout.

## Decision Checklist

- The feature goal is stated without naming a solution (not "we need RAG" upfront).
- Private or changing knowledge requires retrieval or tools — not prompt-only.
- Mutations or live data go through tools with server-side auth — not model memory.
- Fine-tuning is justified only when style/format dominates and eval data exists.
- Sync vs async matches user expectations and timeout budgets.
- Fallback path exists when the primary model or retrieval fails.
- Cost and latency targets are explicit per request type.
- Security and eval plans are named before coding starts.

## Output

Deliver an LLM application architecture brief with:

- **Problem and constraints** — users, data, SLAs, compliance
- **Recommended pattern** — primary approach and why alternatives were rejected
- **Component diagram** — client, API, model gateway, retrieval, tools, storage (text or bullets)
- **Model and context plan** — routing, budgets, memory, citations if applicable
- **Phased rollout** — MVP vs later (eval, rerank, multi-agent, fine-tune)
- **Open questions** — decisions needing product/legal/data input
- **Next skills** — which specialized review to run next
