# Tool Schema and Contracts

## Well-Designed Tool Definition

```json
{
  "name": "search_orders",
  "description": "Search orders for the authenticated customer by status or date range. Do not use for refunds or cancellations.",
  "parameters": {
    "type": "object",
    "properties": {
      "status": {
        "type": "string",
        "enum": ["OPEN", "SHIPPED", "CANCELLED"]
      },
      "fromDate": { "type": "string", "format": "date" },
      "toDate": { "type": "string", "format": "date" }
    },
    "additionalProperties": false
  }
}
```

## Naming Rules

- Verb-first, specific: `search_orders`, not `data` or `helper`
- One responsibility per tool — avoid Swiss-army tools the model mis-selects
- Separate read tools from write tools; do not combine in one function

## Description Quality

Include:

- **Purpose** — what business action it performs
- **Scope limits** — tenant/user context implied by auth, not parameters
- **Negative guidance** — when another tool is correct instead

The model reads descriptions more than parameter names.

## Structured Tool Results

```json
{
  "ok": true,
  "orders": [
    { "id": "ord_123", "status": "SHIPPED", "total": "49.99" }
  ],
  "truncated": false
}
```

On error:

```json
{
  "ok": false,
  "errorCode": "FORBIDDEN",
  "message": "Not authorized to access this order"
}
```

Keep messages safe for model re-prompting — no stack traces or internal IDs unless needed.

## Anti-Patterns

- `run_query(sql: string)` exposed to the model
- Optional `userId` parameter trusting model-supplied identity
- Tools returning full database rows with sensitive columns
- Vague tool named `execute` with free-form input
