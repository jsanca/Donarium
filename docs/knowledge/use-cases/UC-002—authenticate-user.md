UC-002 — Authenticate User
Objetivo

Autenticar a un usuario registrado de Donarium mediante correo y contraseña, emitir un token de sesión firmado y proporcionar el contexto necesario para que el cliente determine la experiencia inicial del usuario.

Este caso de uso verifica identidad. La autorización detallada de cada operación continúa siendo responsabilidad de los casos de uso protegidos.

Actores
Actor	Responsabilidad
Usuario registrado	Proporciona su correo y contraseña.
Donarium	Normaliza el correo, verifica las credenciales, emite el token y devuelve el contexto autenticado.
Cliente web	Valida la entrada para mejorar la experiencia y selecciona la siguiente pantalla según el contexto devuelto.
Precondiciones
Donarium ya fue inicializado.
Existe un User con sus Credentials.
El usuario proporciona un correo y una contraseña.
La contraseña se recibe como RawPassword y permanece únicamente en memoria durante la verificación.
El backend dispone de una clave válida para firmar tokens.
Flujo principal
El usuario abre la pantalla de login.
El cliente valida que el correo y la contraseña estén presentes y que el correo tenga un formato básico válido.
El cliente envía el correo y la contraseña a Donarium mediante una conexión segura.
Donarium normaliza y valida el correo.
Donarium busca el User y sus Credentials mediante el correo normalizado.
Donarium verifica RawPassword contra el PasswordHash.
Donarium carga los PlatformGrant y Membership vigentes del usuario.
Donarium determina el defaultContext del usuario.
Donarium emite un token de sesión firmado y de duración limitada.
Donarium devuelve la identidad autenticada, sus contextos disponibles y el contexto predeterminado.
El cliente descarta la contraseña de su estado.
El cliente determina la siguiente pantalla según defaultContext y los perfiles disponibles.
Flujos alternativos
A1 — Correo con formato inválido

Si el correo no puede normalizarse o no tiene un formato válido, Donarium rechaza la solicitud con:

400 Bad Request

Código:

INVALID_EMAIL

No se realiza una búsqueda de credenciales.

A2 — Credenciales inválidas

Si:

el correo no identifica a un usuario;
el usuario no tiene credenciales;
la contraseña no coincide;

Donarium responde:

401 Unauthorized

Código:

INVALID_CREDENTIALS

La respuesta no revela cuál de las condiciones ocurrió.

No se emite token.

A3 — Solicitud incompleta

Si falta el correo o la contraseña, Donarium responde:

400 Bad Request

con un error de validación consistente.

A4 — Usuario sin contexto de acceso

Si las credenciales son correctas pero el usuario no posee ninguna Membership ni PlatformGrant que le permita acceder, Donarium rechaza el login con una respuesta genérica de acceso no autorizado.

Para evitar enumeración de estados internos, el cliente no recibe detalles sensibles sobre permisos o configuración.

A5 — Fallo interno

Si ocurre un fallo inesperado al consultar credenciales, cargar contextos o emitir el token, Donarium responde:

500 Internal Server Error

No expone detalles de infraestructura.

A6 — Token expirado en una solicitud posterior

Cuando una solicitud protegida presenta un token expirado, Donarium responde:

401 Unauthorized

Código:

SESSION_EXPIRED

El cliente elimina el estado autenticado y muestra nuevamente el login.

Postcondiciones de éxito
La contraseña original no queda persistida, registrada ni incluida en la respuesta.
Se emite un token firmado para el usuario autenticado.
El token tiene una expiración limitada.
El backend no almacena una sesión de servidor.
El cliente recibe los contextos necesarios para seleccionar la siguiente experiencia.
El usuario puede realizar solicitudes autenticadas hasta que el token expire.
Postcondiciones de fallo
No se emite ningún token.
No se modifica el User, sus Credentials, Memberships ni PlatformGrants.
No se revela si el correo está registrado.
El intento puede quedar registrado mediante metadata segura para observabilidad.
Reglas de negocio propuestas
BR-10 — Autenticación por email y password
BR-11 — Validación autoritativa en backend
BR-12 — Error de credenciales no revelador
BR-13 — Password tratada como dato sensible durante login
BR-14 — Token de sesión firmado y de duración limitada
BR-15 — Backend de sesión stateless
BR-16 — Contexto autenticado separado de navegación
BR-17 — Sin bloqueo permanente por intentos fallidos
BR-18 — Autorización basada en PlatformGrant y Membership
Modelo conceptual
LoginRequest
    │
    ├── Email
    └── RawPassword
           │
           ▼
      Authentication
           │
           ├── User
           ├── Credentials ── PasswordHash
           ├── PlatformGrant
           └── Membership
                    │
                    ▼
       AuthenticatedPrincipal
           │
           ├── SessionToken
           ├── PlatformRoles
           ├── OrganizationContexts
           └── DefaultContext
