# TC-001 — Initialize Donarium

## Identificación

| Campo | Valor |
| --- | --- |
| Test Case ID | TC-001 |
| Use Case | [UC-001 — Initialize Donarium](../use-cases/UC-001-initialize-donarium.md) |
| Feature | Initialize Donarium |
| Test Type | Functional / API / Persistence |
| Priority | Critical |
| Estado | Diseñado; no ejecutado |

## Objetivo

Validar que una instalación nueva pueda crear atómicamente un User, Credentials,
Organization, Membership con `OrganizationRole.OWNER` y PlatformGrant con
`PlatformRole.SUPER_ADMIN`. Validar también que el flujo rechace entradas
inválidas o una segunda inicialización sin alterar el estado persistido.

Este documento define comportamiento observable para QA. No presupone que cada
escenario deba implementarse como una prueba automatizada concreta.

## Ambiente y precondiciones

- El servidor de pruebas está iniciado en `DONARIUM_TEST_BASE_URL`; usar
  `http://127.0.0.1:18080` como valor de referencia, nunca el servidor de
  desarrollo de una persona.
- PostgreSQL usa una base dedicada identificada por `TEST_DATABASE_URL`, por
  ejemplo una base exclusiva `donarium_tc_001`. Está prohibido usar una base de
  desarrollo, compartida o con datos reales.
- Las migraciones están aplicadas a esa base de pruebas.
- Antes de cada escenario independiente, la base está en estado no inicializado
  y no contiene datos seed de User, Credential, Organization, Membership ni
  PlatformGrant.
- La persona ejecutora puede inspeccionar las cinco tablas mediante una conexión
  de solo prueba al PostgreSQL dedicado, o mediante una herramienta de consulta
  aprobada para ese entorno.
- La salida de logs del servidor de pruebas se captura para los escenarios que
  verifican no registrar contraseñas.

## Contrato HTTP actual

| Operación | Método y path | Éxito | Error actual |
| --- | --- | --- | --- |
| Estado de setup | `GET /api/setup/status` | `200`, `{"initialized": boolean}` | `500`, `{"error":"internal server error"}` |
| Inicializar | `POST /api/setup` | `201`, `{"userId":"…","organizationId":"…"}` | `400`, `409` o `500`, `{"error":"…"}` |

Las respuestas JSON usan `Content-Type: application/json`. El contrato actual
expone un campo textual `error`; no existe un campo de código de error separado.

## Datos de prueba base

```json
{
  "displayName": "Maria Owner",
  "organizationName": "Northstar Rentals",
  "organizationSlug": "northstar-rentals",
  "email": "OWNER@Northstar.Example ",
  "password": "SecurePass1!"
}
```

El email esperado después de normalizarse es `owner@northstar.example`.

Para la segunda inicialización usar datos diferentes:

```json
{
  "displayName": "Ana Second",
  "organizationName": "Second Rentals",
  "organizationSlug": "second-rentals",
  "email": "ana@second.example",
  "password": "SecurePass1!"
}
```

## Escenarios

### TC-001-01 — Estado inicial no configurado

**Precondición:** Base dedicada vacía y migrada.

**Acción:** Enviar `GET /api/setup/status`.

**Resultados esperados:**

- HTTP `200`.
- `Content-Type: application/json`.
- JSON válido con `{"initialized": false}`.
- No se crea ninguna fila en las cinco tablas de Identity.

### TC-001-02 — Inicialización exitosa

**Precondición:** Base dedicada no inicializada.

**Acción:** Enviar `POST /api/setup` con los datos de prueba base.

**Resultados HTTP esperados:**

- HTTP `201`.
- `Content-Type: application/json`.
- JSON válido con `userId` y `organizationId` no vacíos y con formato UUID.
- La respuesta no contiene `password`, `passwordHash`, `hash` ni el valor
  `SecurePass1!`.

**Resultados funcionales y de persistencia esperados:**

- Existe exactamente un User con email `owner@northstar.example` y nombre
  `Maria Owner`.
- Existe exactamente un Credential para ese User. Su valor almacenado es un
  hash Argon2id, no coincide con `SecurePass1!` ni lo contiene en texto plano.
- Existe exactamente una Organization `Northstar Rentals` con slug
  `northstar-rentals`.
- Existe exactamente una Membership entre ese User y esa Organization con
  `role = owner`.
- Existe exactamente un PlatformGrant para ese User con `role = super_admin`.
- No existen filas adicionales en las cinco tablas.

### TC-001-03 — Estado después de inicializar

**Precondición:** TC-001-02 se completó exitosamente sin limpiar la base.

**Acción:** Enviar `GET /api/setup/status`.

**Resultados esperados:**

- HTTP `200`.
- `Content-Type: application/json`.
- JSON válido con `{"initialized": true}`.

### TC-001-04 — Segunda inicialización

**Precondición:** TC-001-02 se completó exitosamente.

**Acción:** Enviar un segundo `POST /api/setup` con los datos alternativos.

**Resultados esperados:**

- HTTP `409`.
- `Content-Type: application/json`.
- JSON válido con `{"error":"system is already initialized"}` conforme al
  contrato actual.
- Los conteos de User, Credential, Organization, Membership y PlatformGrant
  permanecen en uno.
- El email, slug, roles y hashes creados por TC-001-02 no cambian.

### TC-001-05 — Email inválido

**Precondición:** Base dedicada no inicializada.

**Acción:** Enviar el payload base cambiando `email` a `not-an-email`.

**Resultados esperados:**

- HTTP `400`.
- `Content-Type: application/json`.
- JSON válido con `{"error":"email is not valid"}`.
- Los conteos de las cinco tablas permanecen en cero.

### TC-001-06 — Password débil

**Precondición:** Base dedicada no inicializada y captura de logs activa.

**Acción:** Enviar el payload base cambiando `password` a `weak`.

**Resultados esperados:**

- HTTP `400`.
- `Content-Type: application/json`.
- JSON válido con `{"error":"password does not meet requirements"}`.
- Los conteos de las cinco tablas permanecen en cero.
- La salida de logs no contiene el valor exacto `weak` ni el campo password.

### TC-001-07 — Campos requeridos ausentes

**Precondición:** Base dedicada no inicializada.

**Acción:** Enviar variantes del payload base omitiendo un campo a la vez.

| Variante omitida | Error esperado |
| --- | --- |
| `displayName` | `displayName is required` |
| `email` | `email is required` |
| `password` | `password is required` |
| `organizationName` | `organizationName is required` |
| `organizationSlug` | `organizationSlug is required` |

**Resultados esperados para cada variante:**

- HTTP `400` y `Content-Type: application/json`.
- DTO JSON con el campo `error` y el mensaje esperado de la tabla.
- Los conteos de las cinco tablas permanecen en cero.

### TC-001-08 — Método HTTP incorrecto

**Precondición:** No depende del estado de inicialización.

**Acción:** Enviar `GET /api/setup`. Como variante de contrato, enviar también
`POST /api/setup/status`.

**Resultados esperados:**

- HTTP `405`.
- Header `Allow` con `POST` para `/api/setup` y `GET` para
  `/api/setup/status`.
- `Content-Type: application/json`.
- JSON válido con `{"error":"method not allowed"}`.
- No hay escrituras.

### TC-001-09 — Rollback posterior a una escritura

**Precondición:** Base dedicada no inicializada.

**Mecanismo seguro de inducción:** Ejecutar una composición de prueba aislada
que use el Transactional Setup Service real y un colaborador de persistencia de
solo prueba que falle de forma determinista después de que al menos User y
Credential hayan sido aceptados para escritura. El fallo no se ejecuta contra
el servidor ni la base de desarrollo compartidos.

**Acción:** Ejecutar el setup con los datos base a través de esa composición de
prueba y provocar el fallo posterior a la primera escritura.

**Resultados esperados:**

- El resultado del setup es un fallo interno controlado; si se ejerce mediante
  HTTP, la respuesta es `500` con `{"error":"internal server error"}`.
- Después de terminar la transacción, los conteos son cero para User,
  Credential, Organization, Membership y PlatformGrant.
- No hay password ni hash en el error o los logs capturados.

### TC-001-10 — Persistencia después de reiniciar

**Precondición:** TC-001-02 se completó exitosamente.

**Acción:** Detener el servidor de pruebas, iniciarlo otra vez contra la misma
base dedicada y enviar `GET /api/setup/status`; después enviar otro
`POST /api/setup` con los datos alternativos.

**Resultados esperados:**

- El status devuelve HTTP `200` y `{"initialized": true}`.
- Los cinco registros originales siguen presentes y conservan su relación y
  roles.
- El nuevo setup devuelve HTTP `409` sin modificar esos registros.

Este escenario queda documentado para su ejecución posterior; no se ejecuta
como parte del diseño de TC-001.

### TC-001-11 — Inicializaciones concurrentes

**Clasificación:** Known Technical Debt / Deferred Scenario.

La [ADR-002 — Application-Level Bootstrap Invariant](../decisions/ADR-002-application-level-bootstrap-invariant.md)
acepta una ventana teórica de carrera para dos solicitudes de setup exactamente
concurrentes. Este ciclo no exige implementar ni marcar como fallido un test de
concurrencia. Si el modelo de despliegue evoluciona, el escenario debe
reactivarse con dos payloads válidos de email y slug distintos y una expectativa
actualizada de unicidad persistente.

## Matriz de trazabilidad

| Regla / flujo | Escenarios TC-001 | Estado de cobertura |
| --- | --- | --- |
| Flujo principal | 02, 03, 10 | Diseñado |
| A1 — instalación ya inicializada | 04, 10 | Diseñado |
| A2 — email inválido o ya existente | 05 | Diseñado para email inválido; email duplicado se valida por inspección del índice único, porque A1 intercepta un segundo setup normal |
| A3 — contraseña no conforme | 06 | Diseñado |
| A4 — fallo durante creación | 09 | Diseñado para ejecución aislada posterior |
| BR-01 — inicialización única | 02, 04, 10 | Diseñado para ejecución secuencial; 11 diferido por ADR-002 |
| BR-02 — email normalizado | 02, 05 | Diseñado |
| BR-03 — email único | 02, 04 | Validado por inspección de constraint; no hay caso HTTP independiente alcanzable tras BR-01 |
| BR-04 — password sensible | 02, 06, 09 | Diseñado: respuesta, persistencia y logs |
| BR-05 — roles contextuales | 02 | Diseñado: Membership OWNER y PlatformGrant SUPER_ADMIN |
| BR-06 — transacción | 02, 09 | Diseñado |
| BR-07 — PasswordPolicy | 02, 06 | Diseñado |
| BR-08 — primer usuario | 02 | Diseñado |
| BR-09 — sin usuarios seed | 01 | Validado por precondición e inspección de la base dedicada |

## Evidencia de ejecución

TC-001 es un diseño de prueba. No se ejecutaron requests, servidor, PostgreSQL
ni pruebas automatizadas al crear este documento.
