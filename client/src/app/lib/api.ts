export type OrganizationContext = {
  organizationId: string
  organizationName: string
  organizationSlug: string
  role: string
}

export type DefaultContext = {
  type: string
  organizationId: string
  role: string
}

export type AuthenticatedPrincipal = {
  userId: string
  displayName: string
  email: string
  platformRoles: string[]
  organizationContexts: OrganizationContext[]
  defaultContext: DefaultContext
}

type LoginPrincipal = AuthenticatedPrincipal

export type ApiError = {
  status: number
  message: string
  allow?: string | null
}

function parseErrorBody(status: number, body: unknown): string {
  if (body && typeof body === 'object' && 'error' in body) {
    const err = (body as { error: unknown }).error
    if (typeof err === 'string' && err.length > 0) return err
  }
  return `Unexpected response (${String(status)}).`
}

export async function parseApiError(res: Response): Promise<ApiError> {
  let message = `Unexpected response (${String(res.status)}).`
  const allow = res.headers.get('Allow')
  const ct = res.headers.get('Content-Type') ?? ''
  if (ct.includes('application/json')) {
    try {
      const body: unknown = await res.json()
      message = parseErrorBody(res.status, body)
    } catch {
      // keep default
    }
  } else {
    try {
      const text = await res.text()
      if (text) message = text
    } catch {
      // keep default
    }
  }
  return { status: res.status, message, allow }
}

export async function fetchMe(signal?: AbortSignal): Promise<AuthenticatedPrincipal> {
  const res = await fetch('/api/auth/me', {
    method: 'GET',
    credentials: 'include',
    signal,
    headers: { Accept: 'application/json' },
  })
  if (!res.ok) throw await parseApiError(res)
  const data = (await res.json()) as { principal: LoginPrincipal }
  return data.principal
}

export async function loginRequest(email: string, password: string): Promise<AuthenticatedPrincipal> {
  const res = await fetch('/api/auth/login', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  if (!res.ok) throw await parseApiError(res)
  const data = (await res.json()) as { principal: LoginPrincipal }
  return data.principal
}

export type PropertiesListItem = {
  id: string
}

export async function fetchProperties(signal?: AbortSignal): Promise<PropertiesListItem[]> {
  const res = await fetch('/api/properties', {
    method: 'GET',
    credentials: 'include',
    signal,
    headers: { Accept: 'application/json' },
  })
  if (!res.ok) throw await parseApiError(res)
  // EP-001.02 shape is { properties: [...] }; tolerate alternative shapes for
  // forward compatibility (plain array, {data}, {items}).
  const body: unknown = await res.json()
  if (Array.isArray(body)) return body as PropertiesListItem[]
  if (body && typeof body === 'object') {
    const obj = body as Record<string, unknown>
    if (Array.isArray(obj.properties)) return obj.properties as PropertiesListItem[]
    if (Array.isArray(obj.data)) return obj.data as PropertiesListItem[]
    if (Array.isArray(obj.items)) return obj.items as PropertiesListItem[]
  }
  return []
}
