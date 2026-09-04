# RAG Pipeline and Retrieval

## Reference Pipeline

```text
Source (CMS, DB, files)
  -> Extract text + metadata
  -> Chunk (structure-aware)
  -> Embed
  -> Vector store (+ optional keyword index)
  -> Query: embed + filter + hybrid merge
  -> Optional rerank
  -> Prompt assembly with citations
  -> LLM generate
  -> Audit log (query, sources, latency, tokens)
```

## Chunking Guidelines

| Content type | Approach |
| --- | --- |
| Markdown/docs | Split on headings; keep section path in metadata |
| Product catalog | One product or variant per chunk; stable IDs |
| API docs | Per endpoint or operation with method + path metadata |
| Chat logs | Usually exclude or heavily redact — poor RAG corpus |

Target chunk size to embedding model context — often 300–800 tokens with overlap only when sentences span boundaries.

## Hybrid Retrieval Pattern

Combine lexical + vector when users use exact SKUs, error codes, or product names.

```text
vector_hits = top_k_vector(query, filter=tenant_id)
keyword_hits = bm25(query, filter=tenant_id)
merged = reciprocal_rank_fusion(vector_hits, keyword_hits)
reranked = cross_encoder_rerank(query, merged[:20])[:5]
```

## Metadata Filters (Required for Multi-Tenant)

```json
{
  "tenant_id": "acme",
  "doc_type": "product",
  "locale": "en-us",
  "visibility": "public"
}
```

Apply filters **before** top-k selection. Never rely on the LLM to ignore other tenants' chunks.

## Grounding Prompt Shape

```text
Answer using ONLY the provided context.
If the context is insufficient, say you do not know.
Cite sources as [1], [2] matching the context list.

Context:
[1] ...
[2] ...
```

## Common Failure Modes

- **Stale index** — published content not re-embedded
- **Wrong chunk** — overlap splits tables/code across chunks
- **Over-fetch** — top-k too high; noise drowns signal
- **Under-fetch** — top-k too low; misses correct doc
- **No ACL on retrieval** — cross-tenant leakage risk
