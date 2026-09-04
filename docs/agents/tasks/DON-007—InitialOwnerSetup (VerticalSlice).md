DON-007 — Initial Owner Setup (Vertical Slice)

Status: PLANNED

Owner: Deep Pro

Role: Backend Engineer

Target: dividir en varios sub-slices

No implementarlo como una sola tarea.

DON-007.1

Identity Domain

Crear únicamente el dominio:

User

Credential

Organization

Membership

Roles

PasswordPolicy

Repositories

Services interfaces

Sin PostgreSQL.

Sin HTTP.

Sin pgx.

Sin handlers.

DON-007.2

Persistence

Implementar:

Migraciones

Tablas

Repositories pgx

Constraints

Unique

Indexes

Todavía sin endpoint.

DON-007.3

Transactional InitialOwnerService

Implementar:

Canonical Service

Transactional Decorator

TransactionManager

Password hashing

Email normalization

Aquí adoptamos la decisión de hoy:

Los repositorios reciben el executor (tx o conn) como parámetro.

No usar context.Context para esconder la transacción.

DON-007.4

REST

Implementar:

POST /api/setup

DTO

Validation

HTTP mapping

Problem responses

Todavía sin frontend.

DON-007.5

End-to-End

Probar:

Nueva instalación

↓

POST /setup

↓

Usuario creado

↓

Organization creada

↓

Membership OWNER

↓

Credential Argon2id

↓

Instalación inicializada

Agregar pruebas de:

email duplicado
password inválido
instalación ya inicializada
rollback
