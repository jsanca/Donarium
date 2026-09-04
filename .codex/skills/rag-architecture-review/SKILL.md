---
name: rag-architecture-review
description: Review retrieval-augmented generation systems for indexing, retrieval quality, grounding, evaluation, cost, and operations. Use when designing or reviewing RAG pipelines, vector search, hybrid retrieval, re-ranking, or production chat over private knowledge bases.
---

# RAG Architecture Review

## Workflow

1. Map the RAG pipeline: ingest → chunk → embed → index → retrieve → (rerank) → prompt → generate → audit.
2. Identify data sources, freshness requirements, access control boundaries, and tenant isolation needs.
3. Evaluate retrieval quality drivers: chunking, metadata filters, embeddings model, hybrid search, and top-k strategy.
4. Review grounding and safety: citations, refusal when context insufficient, prompt injection from documents, and PII in corpus.
5. Assess evaluation: offline datasets, golden questions, retrieval metrics, and human review loops.
6. Review operations: reindex strategy, embedding model changes, latency, cost, and observability per stage.
7. Recommend smallest improvements with measurable retrieval or answer-quality impact.

## References

- Read `references/pipeline-and-retrieval.md` for chunking, hybrid search, filtering, and reranking patterns.
- Read `references/evaluation-and-operations.md` for eval sets, metrics, reindexing, and production monitoring.

## Review Checklist

- Documents are chunked with structure-aware boundaries where possible.
- Metadata supports filtering (tenant, product, locale, version, ACL).
- Retrieval returns cite-able sources linked in the answer.
- Access control enforced at retrieval time — not only in the UI.
- Eval set exists with expected sources and acceptance criteria.
- Ingestion is idempotent and supports updates/deletes.
- Latency and token budgets are measured per request.
- Failure modes degrade gracefully (empty retrieval, model timeout).

## Output

Deliver a RAG architecture review with:

- **Pipeline diagram** (text or bullets) and data flows
- **Findings** — severity, component, impact on quality/cost/safety
- **Recommendations** — ordered quick wins vs structural changes
- **Eval plan** — questions, metrics, and pass thresholds
- **Operational gaps** — reindex, versioning, on-call signals
