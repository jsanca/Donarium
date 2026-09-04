# REST Contract Patterns

## Resource-Oriented Paths

```http
GET    /api/v1/orders/{orderId}
GET    /api/v1/orders?page=0&size=20&status=OPEN
POST   /api/v1/orders
PATCH  /api/v1/orders/{orderId}
POST   /api/v1/orders/{orderId}/cancel
```

Prefer sub-resources and actions only when they are not natural CRUD on the parent.

## Status Codes (Common Set)

| Code | Use |
| --- | --- |
| `200` | Successful read or update with body |
| `201` | Resource created |
| `204` | Success with no body (delete, idempotent noop) |
| `400` | Client validation error |
| `401` | Unauthenticated |
| `403` | Authenticated but not authorized |
| `404` | Resource not found (or hidden for auth reasons — be consistent) |
| `409` | Conflict (duplicate, state transition invalid) |
| `422` | Semantic validation failed (optional if distinct from 400) |
| `429` | Rate limited |
| `500` | Unexpected server error (no internal details in body) |

## Idempotency for Writes

```http
POST /api/v1/payments
Idempotency-Key: 7b3e2f1a-4c8d-4e5a-9b2c-1d4e5f6a7b8c
Content-Type: application/json

{ "orderId": "...", "amount": "49.99", "currency": "USD" }
```

Repeat requests with the same key must return the same result without duplicate side effects.

## Pagination Response Shape

```json
{
  "items": [ { "id": "...", "status": "OPEN" } ],
  "page": 0,
  "size": 20,
  "totalItems": 153,
  "totalPages": 8
}
```

Document max `size`, sort fields, and filter semantics. Prefer cursor pagination for very large datasets.
