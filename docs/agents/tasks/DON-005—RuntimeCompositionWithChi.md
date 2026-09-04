DON-005 — Runtime Composition with Chi

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

docs/agents/checkpoints/CHECKPOINT-DON-005-runtime-composition-with-chi.md
Contexto

DON-004 dejó un runtime mínimo funcional con:

carga de configuración;
conexión PostgreSQL mediante pgxpool;
servidor HTTP;
/health/live;
/health/ready;
graceful shutdown;
log/slog;
Docker Compose.

La decisión arquitectónica para Donarium es:

main
↓
application runtime / composite runtime
↓
module runtimes

Además, el router estándar será:

github.com/go-chi/chi/v5

No continuar expandiendo el routing directo con http.ServeMux.

Objetivo

Refactorizar el bootstrap actual para:

adoptar Chi como router;
introducir composición explícita de runtimes;
mantener main.go como composition root;
encapsular health dentro de un runtime de plataforma;
preparar el límite natural para futuros runtimes de dominio.

El comportamiento observable de health no debe cambiar.

Arquitectura objetivo
cmd/donarium/main.go
        │
        ▼
platform/runtime.ApplicationRuntime
        │
        ├── PlatformRuntime
        │      └── health routes
        │
        └── future module runtimes
               ├── IdentityRuntime
               ├── PropertiesRuntime
               ├── LeasesRuntime
               └── ...

main.go debe limitarse a:

cargar configuración;
crear dependencias compartidas;
crear runtimes;
componer la aplicación;
ejecutar y manejar errores finales.

No debe registrar rutas directamente.

Alcance
1. Introducir Chi

Agregar:

github.com/go-chi/chi/v5

Usar Chi como router principal.

Ejemplo conceptual:

router := chi.NewRouter()

No introducir todavía todos los middlewares de Chi por defecto.

Puede utilizar únicamente los que aporten valor inmediato y no cambien innecesariamente el comportamiento existente.

No agregar todavía:

CORS;
autenticación;
rate limiting;
compresión;
request logging completo;
tracing.
2. Contrato de runtime modular

Definir un contrato pequeño y estable para que cada módulo registre su superficie HTTP.

Una forma aceptable:

type ModuleRuntime interface {
    RegisterRoutes(router chi.Router)
}

También puede utilizarse otro nombre si expresa mejor la intención, pero evitar contratos genéricos ambiguos como:

Start()
Run()
Init()
Setup()

cuando la responsabilidad real en este slice es registrar composición HTTP.

No agregar lifecycle hooks que todavía no sean necesarios.

3. Application Runtime

Crear un runtime de aplicación responsable de:

poseer el router Chi;
registrar los runtimes de módulo;
construir o encapsular el http.Server;
exponer el ciclo de ejecución;
realizar graceful shutdown con las dependencias necesarias.

Estructura sugerida:

server/internal/platform/runtime/
├── application.go
├── module.go
└── application_test.go

La nomenclatura puede ajustarse si queda más clara.

Ejemplo conceptual:

appRuntime := runtime.NewApplication(
    cfg,
    logger,
    platformRuntime,
)

if err := appRuntime.Run(ctx); err != nil {
    ...
}

No convertirlo en un framework interno.

4. Platform Runtime

Crear un runtime específico de plataforma:

server/internal/platform/runtime/platform.go

o equivalente.

Debe ser responsable de registrar:

GET /health/live
GET /health/ready

El runtime de plataforma recibe las dependencias que necesita, por ejemplo:

logger;
readiness checker;
configuración de health, si existe.

No debe abrir conexiones por sí mismo si estas ya fueron creadas por el composition root.

5. Main como composition root

Refactorizar:

server/cmd/donarium/main.go

para que no contenga:

creación manual del mux;
registro directo de health;
detalles internos de handlers;
construcción detallada del servidor HTTP.

Debe seguir siendo responsable de crear dependencias concretas:

Config
Logger
Postgres pool
PlatformRuntime
ApplicationRuntime

Esto es composition root, no service locator.

Límite importante

No crear todavía runtimes vacíos para todos los dominios.

Evitar archivos como:

identity/runtime.go
properties/runtime.go
leases/runtime.go

sin comportamiento real.

El contrato queda preparado, pero cada runtime de dominio se creará cuando llegue su primer slice vertical.

Este slice debe incluir solamente:

ApplicationRuntime
PlatformRuntime
Health

Preservar exactamente los contratos existentes:

GET /health/live
{
  "status": "ok"
}
GET /health/ready

Disponible:

{
  "status": "ready",
  "checks": {
    "database": "up"
  }
}

No disponible:

{
  "status": "not_ready",
  "checks": {
    "database": "down"
  }
}

Mantener:

200 para live;
200 para ready saludable;
503 para ready no saludable;
405 para métodos no permitidos;
Content-Type: application/json;
ausencia de detalles internos.
Configuración

No ampliar innecesariamente Config.

El Config actual puede mantenerse plano en este slice.

Solo modificarlo si la composición del runtime exige una mejora concreta.

No introducir todavía:

Viper;
Koanf;
archivos YAML;
configuración por módulo;
hot reload.
Shutdown y lifecycle

Mantener el flujo:

SIGINT / SIGTERM
↓
HTTP shutdown
↓
database pool close
↓
process exit

La propiedad de los recursos debe quedar clara:

main crea el pool;
el runtime usa el pool;
el cierre ocurre una sola vez;
no debe existir doble Close().

Puede resolverse dejando que:

ApplicationRuntime cierre el servidor HTTP;
main cierre PostgreSQL;

o pasando closers explícitos al runtime.

Elegir una opción simple y documentarla.

Pruebas

Mantener las pruebas existentes de health.

Agregar pruebas para composición.

Como mínimo:

Application runtime
registra rutas de los módulos entregados;
un runtime fake puede registrar una ruta de prueba;
no depende de PostgreSQL real;
no requiere levantar un puerto real si httptest es suficiente.
Platform runtime
registra ambas rutas de health;
conserva respuestas y códigos actuales.

No usar mocks framework-based.

Ejemplo de fake:

type fakeRuntime struct {
    registered bool
}

func (f *fakeRuntime) RegisterRoutes(router chi.Router) {
    f.registered = true
    router.Get("/test", ...)
}
Dependencias

Actualizar:

server/go.mod
server/go.sum

Ejecutar:

go mod tidy

No agregar ninguna otra dependencia externa salvo que sea indispensable.

Validación

Ejecutar y documentar:

go fmt ./...
go vet ./...
go test ./...
golangci-lint run
go build ./cmd/donarium
docker compose config
docker compose up -d --build
curl -i http://localhost:${HTTP_PORT}/health/live
curl -i http://localhost:${HTTP_PORT}/health/ready
docker compose stop postgres
curl -i http://localhost:${HTTP_PORT}/health/live
curl -i http://localhost:${HTTP_PORT}/health/ready
docker compose down

Resultados esperados:

DB up:
live  → 200
ready → 200

DB down:
live  → 200
ready → 503

Confirmar que no queden contenedores ejecutándose al finalizar.

Restricciones

No implementar:

autenticación;
identity;
sesiones;
JWT;
cookies;
usuarios;
organizaciones;
CORS;
middleware de logging completo;
request ID;
recovery middleware, salvo que exista una razón concreta y documentada;
DI framework;
event bus;
lifecycle framework;
runtime registry dinámico;
reflection para descubrir módulos.

No modificar:

client/**
knowledge/**
README.md
ROADMAP.md
AGENTS.md
Entregables
server/cmd/donarium/main.go

server/internal/platform/runtime/**
server/internal/platform/http/**
server/internal/platform/http/health/**

server/go.mod
server/go.sum

docs/agents/tasks/DON-005-runtime-composition-with-chi.md
docs/agents/reports/DON-005-runtime-composition-with-chi.md
docs/agents/checkpoints/CHECKPOINT-DON-005-runtime-composition-with-chi.md

Los archivos existentes pueden moverse cuando la nueva organización sea más clara, pero evitar renombrados cosméticos no necesarios.

Definition of Done
Chi es el router principal.
No se usa http.ServeMux como router de aplicación.
Existe un contrato pequeño para runtimes modulares.
Existe ApplicationRuntime.
Existe PlatformRuntime.
Health se registra desde PlatformRuntime.
main.go funciona como composition root.
main.go no registra rutas directamente.
No existen runtimes vacíos de dominio.
Los contratos HTTP de health no cambian.
Graceful shutdown sigue funcionando.
Build, tests, vet y lint pasan.
Docker Compose sigue funcionando.
No se modifica client/.
No se implementa lógica de negocio.
No se hacen commits.
Se actualizan checkpoint y reporte.
