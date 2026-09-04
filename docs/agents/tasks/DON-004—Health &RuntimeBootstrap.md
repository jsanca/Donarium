DON-004 — Health & Runtime Bootstrap

Status: COMPLETE
Owner: Deep Pro
Role: Backend Engineer — Go
Target: 30–40 minutos
Hard stop: 50 minutos

Aplicar:

.claude/skills/execution-timebox/SKILL.md
.claude/skills/engineering-reporting/SKILL.md

No commits.

Crear checkpoint:

docs/agents/checkpoints/CHECKPOINT-DON-004-health-runtime-bootstrap.md
Contexto

DON-003 dejó preparado:

monolito modular bajo server/;
módulo Go único;
PostgreSQL mediante Docker Compose;
Dockerfile multi-stage;
Makefile;
configuración de lint;
estructura de módulos y plataforma.

Ahora necesitamos convertir el cascarón en un servidor mínimo ejecutable.

Deep Pro mantiene ownership sobre:

server/**
compose.yaml
Makefile
.env.example
docs/agents/**

No modificar:

client/**
knowledge/**
README.md
ROADMAP.md
AGENTS.md
Objetivo

Implementar el primer runtime del backend de Donarium con:

main.go
→ carga de configuración
→ conexión PostgreSQL
→ servidor HTTP
→ /health/live
→ /health/ready
→ cierre ordenado

Sin implementar todavía:

login;
usuarios;
sesiones;
organizaciones;
lógica de negocio;
migraciones funcionales;
frontend integration.
Endpoints
Liveness
GET /health/live

Debe responder cuando el proceso está vivo, independientemente del estado de PostgreSQL.

Respuesta exitosa:

{
  "status": "ok"
}

Código:

200 OK
Readiness
GET /health/ready

Debe verificar que la aplicación puede atender tráfico y que PostgreSQL responde.

Respuesta exitosa:

{
  "status": "ready",
  "checks": {
    "database": "up"
  }
}

Código:

200 OK

Cuando PostgreSQL no esté disponible:

{
  "status": "not_ready",
  "checks": {
    "database": "down"
  }
}

Código:

503 Service Unavailable

No exponer:

URLs de conexión;
contraseñas;
errores SQL internos;
stack traces;
detalles sensibles.
Estructura esperada

Mantener una estructura mínima dentro de platform.

server/
├── cmd/
│   └── donarium/
│       └── main.go
├── internal/
│   └── platform/
│       ├── config/
│       │   └── config.go
│       ├── database/
│       │   └── postgres.go
│       └── http/
│           ├── server.go
│           └── health/
│               ├── handler.go
│               └── response.go

Puede ajustar nombres cuando exista una razón clara, pero debe conservar las responsabilidades separadas.

No crear aún:

domain/
application/
ports/
adapters/
repository/
service/

para health. Este slice no necesita Clean Architecture ceremonial.

Configuración

Cargar desde variables de entorno.

Como mínimo:

HTTP_PORT
POSTGRES_HOST
POSTGRES_PORT
POSTGRES_DB
POSTGRES_USER
POSTGRES_PASSWORD
POSTGRES_SSLMODE
DATABASE_CONNECT_TIMEOUT
SHUTDOWN_TIMEOUT

Actualizar:

.env.example

con valores locales seguros.

Requisitos:

valores por defecto razonables para desarrollo;
validación de campos obligatorios;
errores claros al iniciar;
no usar una biblioteca pesada si os.Getenv y una estructura pequeña bastan;
no incluir secretos en logs.

Evitar configuración global mutable.

PostgreSQL

Usar el driver:

github.com/jackc/pgx/v5

Preferiblemente mediante:

pgxpool

Requisitos:

pool de conexiones;
contexto con timeout al conectar;
Ping inicial;
cierre del pool durante shutdown;
readiness con timeout corto;
URL o configuración construida de manera segura.

No implementar todavía:

repositorios;
transacciones;
migraciones;
queries de dominio;
ORM;
SQL builder.
HTTP server

Usar inicialmente la biblioteca estándar:

net/http

No introducir Gin, Echo, Fiber o Chi para solo dos endpoints.

Requisitos:

timeouts explícitos;
ReadHeaderTimeout;
ReadTimeout;
WriteTimeout;
IdleTimeout;
rutas registradas de manera clara;
respuestas JSON;
Content-Type: application/json;
métodos no permitidos controlados apropiadamente.

No crear middleware genérico todavía, salvo que sea estrictamente necesario.

Graceful shutdown

Manejar al menos:

SIGINT
SIGTERM

Flujo esperado:

signal
→ dejar de aceptar tráfico
→ HTTP shutdown con timeout
→ cerrar pool PostgreSQL
→ terminar proceso

No llamar os.Exit desde paquetes internos.

main debe controlar la salida del proceso.

Logging

Usar:

log/slog

Requisitos mínimos:

inicio del servidor;
puerto;
conexión exitosa a PostgreSQL;
fallo de startup;
inicio y resultado de shutdown.

No registrar:

password;
connection string completa;
body de requests;
datos sensibles.

No implementar todavía una plataforma de observabilidad completa.

Docker Compose

Agregar el servicio del backend al compose.yaml.

Servicios esperados:

postgres
server

El servicio server debe:

construirse desde server/Dockerfile;
depender de PostgreSQL saludable;
recibir configuración mediante variables;
exponer el puerto HTTP configurable;
compartir la red interna;
soportar rebuild normal con Docker Compose.

No agregar todavía:

frontend;
reverse proxy;
Redis;
Keycloak;
OpenTelemetry stack.
Dockerfile

Ahora sí debe ser construible.

Ajustar el Dockerfile existente para:

descargar dependencias aprovechando cache;
compilar cmd/donarium;
producir un binario único;
copiarlo a runtime;
ejecutar como usuario no root;
incluir solo lo necesario;
exponer el puerto documentalmente;
usar versiones explícitas.

No usar:

latest;
UPX;
CGO innecesario;
shell complejo de entrada;
dummy files.
Makefile

Actualizar los comandos existentes.

Como mínimo:

make postgres-up
make postgres-down
make server-build
make server-run
make server-up
make server-down
make health-live
make health-ready
make lint
make test
make build

Puede mantener aliases coherentes.

make health-live y make health-ready deberían usar curl o equivalente y fallar si el endpoint no responde correctamente.

La versión de golangci-lint debe quedar fijada en este slice si aún sigue pendiente del anterior.

Pruebas

Implementar pruebas unitarias para los handlers.

Como mínimo:

Liveness
responde 200;
devuelve JSON correcto;
rechaza métodos no permitidos.
Readiness
responde 200 cuando el checker indica disponibilidad;
responde 503 cuando el checker falla;
no filtra el error interno;
devuelve Content-Type correcto.

Para evitar pruebas acopladas a PostgreSQL real, el handler debería depender de una interfaz mínima, por ejemplo:

type ReadinessChecker interface {
    Check(ctx context.Context) error
}

No crear mocks framework-based. Basta un fake pequeño en _test.go.

Agregar una validación de integración opcional o comando manual contra PostgreSQL real, pero no convertir la suite básica en dependiente de Docker.

Contrato interno mínimo

Mantener la abstracción pequeña.

Algo equivalente a:

type ReadinessChecker interface {
    Check(context.Context) error
}

La implementación PostgreSQL puede delegar a:

pool.Ping(ctx)

No generalizar todavía a un árbol de health checks extensible, registry dinámico o plugin framework.

Validación

Ejecutar y documentar:

go mod tidy
go fmt ./...
go vet ./...
go test ./...
golangci-lint run
go build ./cmd/donarium
docker compose config
docker compose up -d --build
docker compose ps
curl -i http://localhost:${HTTP_PORT}/health/live
curl -i http://localhost:${HTTP_PORT}/health/ready
docker compose stop postgres
curl -i http://localhost:${HTTP_PORT}/health/live
curl -i http://localhost:${HTTP_PORT}/health/ready
docker compose down

Resultado esperado después de detener PostgreSQL:

/health/live  → 200
/health/ready → 503

Confirmar que al finalizar no quedan contenedores ejecutándose.

Restricciones

No implementar:

/login;
/auth;
JWT;
cookies;
sesiones;
usuarios;
migraciones SQL;
OpenAPI;
Swagger;
framework HTTP externo;
DI framework;
métricas;
tracing;
frontend proxy;
hot reload todavía.

No modificar el diseño del login ni el código de Elito.

Entregables
server/cmd/donarium/main.go

server/internal/platform/config/**
server/internal/platform/database/**
server/internal/platform/http/**
server/**/*_test.go

server/go.mod
server/go.sum
server/Dockerfile

compose.yaml
.env.example
Makefile

docs/agents/tasks/DON-004-health-runtime-bootstrap.md
docs/agents/reports/DON-004-health-runtime-bootstrap.md
docs/agents/checkpoints/CHECKPOINT-DON-004-health-runtime-bootstrap.md
Definition of Done
El servidor Go compila.
PostgreSQL se conecta durante startup.
/health/live responde 200.
/health/ready responde 200 con PostgreSQL disponible.
/health/ready responde 503 sin PostgreSQL.
/health/live permanece en 200 sin PostgreSQL.
Existe graceful shutdown.
Se usa slog.
No se filtran secretos ni errores internos.
El servicio server arranca mediante Docker Compose.
Las pruebas pasan.
Lint, vet y build pasan.
golangci-lint queda versionado.
No se implementa lógica de negocio.
No se modifica client/.
No se hacen commits.
Se actualizan checkpoint y reporte.
