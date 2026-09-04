/**
 * Validates and normalizes a returnUrl candidate so the client never
 * performs an open redirect. Only same-origin absolute paths are accepted.
 */
export function sanitizeReturnUrl(raw: string | null): string | null {
  if (!raw) return null
  // Must start with / and not with // (protocol-relative) and no scheme
  if (!raw.startsWith('/')) return null
  if (raw.startsWith('//')) return null
  if (/^[a-zA-Z]+:/.test(raw)) return null
  // Reject CRLF injection
  if (raw.includes('\n') || raw.includes('\r')) return null
  return raw
}

export function getReturnUrlFromSearch(search: string): string | null {
  const params = new URLSearchParams(search)
  return sanitizeReturnUrl(params.get('returnUrl'))
}

export function buildLoginUrlWithReturnUrl(returnUrl: string): string {
  return `/login?returnUrl=${encodeURIComponent(returnUrl)}`
}
