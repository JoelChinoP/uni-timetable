# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

- Usuario principal: estudiantes que necesitan planificar y validar su horario semanal seleccionando cursos y grupos compatibles.
- Audiencia secundaria reducida: personas con rol administrador que registran y mantienen usuarios, cursos, docentes, aulas y horarios.
- Ambos roles pueden acceder al panel; el usuario con rol normal no registra a otros usuarios.

## Product Purpose

Uni Timetable centraliza la planificación de horarios universitarios en un panel único. El éxito del producto es que un estudiante combine cursos, grupos y bloques semanales, identifique conflictos y salga con un horario válido sin hacer cruces manuales.

## Positioning

A diferencia de revisar catálogos, horarios y aulas por separado, Uni Timetable une la selección académica, la visualización semanal y la detección de cruces en una sola experiencia operativa.

## Operating Context

- Aplicación web estática en Svelte, servida desde la raíz del repositorio.
- API Go independiente en `backend/`, con PostgreSQL mediante pgx y SQLC.
- Autenticación exclusiva mediante Google Identity Services, con el ID token validado en Go.
- Uso previsto para una comunidad inicial pequeña de alumnos registrados.

## Capabilities and Constraints

- Acceso autenticado solamente con Google; no se mantienen cuentas ni contraseñas propias.
- Persistencia mínima de personas: correo, nombre visible y rol.
- Roles confirmados: `ADMIN` y `USER`.
- El panel es visible para ambos roles; registrar usuarios está reservado a `ADMIN`.
- Los usuarios normales pueden planificar horarios; los administradores también mantienen registros de usuarios, cursos, docentes, aulas y horarios.
- La interfaz y los mensajes principales se mantienen en español.
- Arquitectura separada confirmada: Svelte/Vite estático, Go `net/http`, SQLC y Supabase Transaction Pooler.

## Brand Commitments

- Nombre del producto: Uni Timetable.
- Voz y contenido principales en español.
- Sin compromisos visuales adicionales confirmados en esta inicialización.

## Evidence on Hand

- Sistema visual y componentes actuales en `src/`.
- Horario académico real 2026-B de Ingeniería de Sistemas embebida en `backend/cmd/seed/2026_B_data.json` (35 cursos, 299 sesiones, extracción oficial).
- Esquema relacional base en `backend/internal/database/schema.sql` y migraciones en `backend/internal/database/migrations/`.
- Documentación operativa y de despliegue en `README.md`.
- No hay testimonios, métricas de uso, contenido legal, estudios de caso ni datos reales confirmados; no se deben inventar.

## Product Principles

- Centralizar en un solo panel las decisiones del horario y sus validaciones.
- Persistir solamente la información personal mínima necesaria para operar.
- Separar claramente la experiencia pública, la autenticación y las operaciones de administración.
- Mantener la arquitectura sencilla y los límites de seguridad en el backend, no en el navegador.
