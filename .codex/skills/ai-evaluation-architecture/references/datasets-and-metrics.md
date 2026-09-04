# Datasets and Metrics

## Golden Dataset Structure

Each case should include:

| Field | Purpose |
| --- | --- |
| `id` | Stable reference in CI logs |
| `input` | User message or task prompt |
| `context_setup` | Tenant, user role, feature flags |
| `expected_sources` | Doc IDs/paths for RAG (optional) |
| `expected_answer` | Reference text or structured fields |
| `acceptance_rubric` | How human/automation judges pass |
| `tags` | `happy`, `refusal`, `adversarial`, `acl`, `edge` |

Target **20–50 cases** for MVP; grow by production failures.

## Eval Layers

```text
Layer 1: Retrieval  -> Did we fetch the right chunks?
Layer 2: Generation -> Is the answer grounded and correct?
Layer 3: E2E task   -> Did the user goal succeed (incl. tools)?
Layer 4: Safety     -> Refusal, PII, injection resistance
```

Run Layer 1 before tuning prompts. Fixing generation when retrieval is wrong wastes time.

## Retrieval Metrics

| Metric | When |
| --- | --- |
| Recall@k | Expected doc in top k |
| MRR | Rank of first relevant hit |
| nDCG | Graded relevance |
| Filter correctness | Wrong tenant/doc type excluded |

## Generation / E2E Metrics

| Metric | When |
| --- | --- |
| Exact match | Structured outputs (JSON, codes) |
| Semantic similarity | Paraphrased correct answers (use carefully) |
| Citation match | Answer claims supported by cited chunks |
| Tool call accuracy | Name + args match golden |
| Human rubric 1–5 | Subjective quality; inter-rater sample |

## Rubric Example (Support Bot)

| Score | Criteria |
| --- | --- |
| 5 | Correct, cited, concise, safe |
| 3 | Mostly correct; minor omission |
| 1 | Wrong fact, missing citation, or unsafe |
| 0 | Should have refused |

## Categories to Include

- **Answerable** from corpus
- **Unanswerable** — should refuse or clarify
- **Ambiguous** — should ask follow-up
- **Adversarial** — injection in user msg or retrieved doc
- **ACL** — user must not see other tenant data
- **Tool** — requires correct API call sequence

## Anti-Patterns

- Only demo questions from the team that built the feature
- Single "vibe check" without recorded expected outcomes
- LLM-as-judge as the only metric with no human spot checks
- Eval only after launch with no baseline
