# Errors, Versioning, and Compatibility

## RFC 9457 Problem Details (Recommended Shape)

```json
{
  "type": "https://api.example.com/problems/order-not-found",
  "title": "Order Not Found",
  "status": 404,
  "detail": "No order exists with the given identifier.",
  "instance": "/api/v1/orders/550e8400-e29b-41d4-a716-446655440000",
  "orderId": "550e8400-e29b-41d4-a716-446655440000"
}
```

Rules:

- Do not expose stack traces, SQL, or internal service names in `detail`.
- Use stable `type` URIs or codes for programmatic client handling.
- Keep extension fields consistent across endpoints.

## Safe vs Breaking Changes

| Safe (usually) | Breaking |
| --- | --- |
| Add optional response field | Remove or rename field |
| Add optional request field | Change field type or meaning |
| Add new endpoint | Change status code for same condition |
| Add enum value with tolerant clients | Remove enum value |
| Add new error code | Tighten validation on existing field |

## Versioning Options

1. **URL prefix** — `/api/v1/`, `/api/v2/` — clearest for large breaks
2. **Header** — `Accept-Version: 2` — useful for minor negotiation
3. **Additive-only policy** — no version bump if clients ignore unknown fields

Pick one primary strategy per product surface and document deprecation timelines.

## Deprecation Pattern

```http
GET /api/v1/legacy-reports
Deprecation: true
Sunset: Sat, 01 Nov 2026 00:00:00 GMT
Link: </api/v2/reports>; rel="successor-version"
```

Track usage metrics before removal. Provide migration notes in changelog and OpenAPI description.

## Contract Test Minimum

- Happy path per endpoint
- Auth required paths return `401`/`403` as documented
- Validation errors match documented problem shape
- Pagination boundaries (empty page, max size)
