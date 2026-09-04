# RAG Evaluation and Operations

## Golden Question Set Template

```markdown
| id | question | expected_source_ids | must_not_contain | notes |
|----|----------|---------------------|------------------|-------|
| q1 | What is the return policy for leather goods? | doc-returns-001 | competitor names | needs citation |
| q2 | SKU LP-100 warranty period? | product-lp-100 | | exact SKU match |
```

Run offline after index or embedding model changes.

## Metrics to Track

| Stage | Metric |
| --- | --- |
| Retrieval | hit@k, MRR, source recall on golden set |
| Generation | citation accuracy, hallucination rate (human or LLM-judge) |
| Production | p95 latency, tokens in/out, cost per conversation |
| Quality | thumbs down rate, escalation to human |

Automate retrieval metrics in CI where feasible; sample production conversations for human review.

## Reindex and Model Migration

When changing embedding model or chunk strategy:

1. Build new index version (`v2`) in parallel
2. Run golden eval comparing v1 vs v2
3. Cut over with feature flag or shadow traffic
4. Keep rollback index until burn-in period ends

Document that embedding space is not compatible across models — full re-embed required.

## Observability Fields (Audit Log)

```json
{
  "conversationId": "...",
  "retrievalQuery": "...",
  "sourceIds": ["doc-1", "product-42"],
  "retrievalLatencyMs": 120,
  "generationLatencyMs": 890,
  "tokenUsage": { "prompt": 2100, "completion": 340 },
  "model": "provider/model-name"
}
```

Do not log full user PII or document bodies in production audit streams unless policy allows and redaction exists.

## Cost Controls

- Cap retrieved chunks and max prompt tokens
- Cache embeddings for stable corpora
- Route simple FAQ to rules/SQL before LLM when deterministic
- Rate limit per user/tenant

Use `java-ai-backend-engineering` for Spring implementation patterns; this skill focuses on architecture and quality gates.
