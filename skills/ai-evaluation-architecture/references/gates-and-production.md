# Release Gates and Production Evaluation

## Offline Gates (Pre-Release)

| Gate | Runs when | Blocks release if |
| --- | --- | --- |
| Smoke eval | Every PR (subset) | Critical cases fail |
| Full golden set | Nightly / pre-deploy | Below threshold |
| Retrieval-only | Index/embed change | Recall@k drops > X% |
| Safety suite | Prompt/policy change | Refusal/injection fail |

Keep PR subset **fast** (<5 min). Full set can be async with deploy hold.

## Threshold Example

```yaml
retrieval:
  recall_at_5: ">= 0.85"
generation:
  citation_match: ">= 0.80"
  human_rubric_mean: ">= 4.0"   # sample of 20 cases
safety:
  injection_block_rate: "== 1.0"
  pii_leak_rate: "== 0"
```

Tune thresholds from baseline — not aspirational day one.

## Online / Production Signals

| Signal | Method |
| --- | --- |
| User thumbs / feedback | Explicit; bias toward negatives |
| Implicit | Abandon rate, re-ask rate, escalation |
| Shadow mode | New model answers logged, not shown |
| Sampled human review | Weekly batch from production |
| Retrieval hit rate | Empty retrieval %, avg k used |

Alert on **drift**: sudden drop in feedback or spike in empty retrieval.

## LLM-as-Judge (Use Carefully)

Good for: ranking two answers, checking citation presence, coarse rubric.

Bad as sole gate: nuanced domain correctness, legal/medical, subtle hallucination.

Mitigations:

- Pin judge model and prompt version
- Calibrate against human labels monthly
- Use for triage, not binary ship/no-ship alone

## Versioning Triggers

Re-run full eval when any of:

- System prompt material change
- Embedding model or chunk strategy change
- Top-k, reranker, or filter logic change
- Primary LLM model swap
- Tool schema or auth rules change

Record **baseline scores** in release notes.

## Minimal Harness (First Sprint)

1. YAML/JSON golden file in repo
2. Script: call API or pipeline stub, compare retrieval IDs + optional answer snapshot
3. CI job on `tests/eval/` — fail on regression
4. README: how to add a case when prod fails

Defer dashboards until golden set stabilizes.

## Relationship to Other Skills

- **`llm-application-architecture`** — defines what to build; this skill defines how to measure it
- **`rag-architecture-review`** — retrieval design details; eval implements recall/citation checks
- **`testing-strategy`** — classic test pyramid; AI eval complements — does not replace unit tests
