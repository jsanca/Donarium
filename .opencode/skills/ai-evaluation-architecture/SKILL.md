---
name: ai-evaluation-architecture
description: Design evaluation architecture for LLM, RAG, and agent features. Use when defining golden datasets, offline vs online eval, retrieval metrics, release gates, regression tests for prompts/models, or reviewing whether an AI feature is measurable before production.
---

# AI Evaluation Architecture

## Workflow

1. Define **task success** in user terms — not "good answer" but observable outcomes (correct SKU, policy cited, ticket resolved, no PII leaked).
2. Split eval layers: retrieval quality, generation quality, end-to-end task success, and safety/refusal behavior.
3. Design **golden datasets**: questions, expected sources/answers, edge cases, adversarial prompts, and tenant/ACL scenarios.
4. Choose metrics per layer: recall@k, MRR, citation match, exact/semantic answer match, tool-call correctness, human rubric scores.
5. Plan **offline eval** (CI, pre-release) vs **online eval** (shadow, sample, human review) and who owns each.
6. Define **release gates**: pass thresholds, blockers vs warnings, and what runs on every PR vs nightly vs pre-deploy.
7. Address model/prompt/index changes: versioned eval sets, regression baselines, and re-embed triggers.
8. Recommend minimal eval harness the team can run this sprint — avoid boiling the ocean.

## References

- Read `references/datasets-and-metrics.md` for golden set design, retrieval vs answer metrics, and rubrics.
- Read `references/gates-and-production.md` for CI gates, online monitoring, LLM-as-judge pitfalls, and drift detection.

## Evaluation Checklist

- Success criteria are written before the first demo.
- Golden set includes happy path, "no answer in corpus," and adversarial/injection cases.
- Retrieval eval is separate from generation eval when RAG is used.
- Expected **sources** (not just answers) are recorded for RAG cases.
- Tool-calling eval covers wrong args, auth failure, and idempotent retries.
- Human review loop exists for subjective quality until automation is trusted.
- Eval runs are reproducible (fixed seeds, pinned models, logged prompts).
- Production monitors quality drift — not only latency and errors.
- Model or embedding changes trigger explicit regression run.

## Output

Deliver an AI evaluation architecture plan with:

- **Success definition** — task-level acceptance criteria
- **Eval layers** — retrieval / generation / E2E / safety
- **Golden dataset spec** — size, categories, ownership, refresh cadence
- **Metrics and thresholds** — what passes release
- **Harness placement** — local, CI, staging, production sampling
- **Regression policy** — when to re-baseline after prompt/model/index changes
- **Gaps** — what to build first vs defer
