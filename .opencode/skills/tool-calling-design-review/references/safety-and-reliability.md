# Tool Safety and Reliability

## Authorization Pattern

```java
public ToolResult searchOrders(SearchOrdersInput input, AgentContext ctx) {
    UUID customerId = ctx.authenticatedCustomerId();
    return orderQueryPort.search(customerId, input.status(), input.dateRange());
}
```

Never accept tenant/user identity solely from tool arguments unless cryptographically bound to the session.

## Human Confirmation Gate

Require confirmation UI or second-step token for:

- Refunds and money movement
- Account deletion
- Bulk exports of user data
- Permission or role changes

```text
Agent proposes cancel_order(ord_123)
  -> UI: "Cancel order #123?" [Confirm] [Reject]
  -> Tool executes only after confirm token returned
```

## Idempotency

```java
public ToolResult createTicket(CreateTicketInput input, AgentContext ctx) {
    return idempotencyStore.execute(ctx.idempotencyKey(), () ->
        ticketService.create(ctx.userId(), input.subject(), input.body()));
}
```

## Loop and Budget Limits

```text
max_tool_calls_per_turn: 8
max_tool_calls_per_session: 50
tool_timeout_ms: 5000
```

Stop gracefully and explain limit reached — do not spin until context overflow.

## Observability

Log per tool invocation:

```json
{
  "sessionId": "...",
  "tool": "search_orders",
  "durationMs": 45,
  "ok": true,
  "errorCode": null
}
```

Alert on spikes in `FORBIDDEN`, timeouts, or repeated identical mutating calls.

## Testing Matrix

| Case | Expected |
| --- | --- |
| Valid input, authorized user | Success result |
| Valid input, wrong tenant | FORBIDDEN |
| Invalid enum/date | VALIDATION_ERROR |
| Duplicate idempotency key | Same result, no double side effect |
| Simulated timeout | Structured error; no partial commit without compensation |
