# Logs, Metrics, and Traces

## Structured Log Example

```json
{
  "timestamp": "2026-06-25T14:03:11.123Z",
  "level": "WARN",
  "message": "Payment provider timeout",
  "service": "orders-api",
  "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
  "spanId": "00f067aa0ba902b7",
  "orderId": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "provider": "payments-gateway",
  "durationMs": 5021
}
```

Avoid logging passwords, tokens, full payment payloads, or regulated personal data.

## Metrics Naming (Prometheus Style)

```text
http_server_requests_seconds_count{method="POST",uri="/api/v1/orders",status="500"}
http_server_requests_seconds_sum{...}
http_server_requests_seconds_bucket{le="0.5", ...}
```

Prefer:

- Low-cardinality labels (`status`, `method`, bounded `uri` template)
- Business counters where useful (`orders_completed_total`)

Avoid unbounded labels (`userId`, raw URL paths with IDs).

## Trace Propagation

- Incoming HTTP: extract W3C `traceparent` or platform equivalent
- Outgoing HTTP: inject context into client calls
- Async/message: serialize trace context in message headers or metadata
- Logs: include `traceId` so logs and traces link in the APM UI

## Blind Spot Questions

- Can on-call tell **which dependency** failed from dashboards alone?
- Can you find **all logs for one request** without grep across raw text files?
- Do you have **before/after deploy** comparison for error rate and latency?
- Are batch jobs and schedulers visible — not only HTTP?
