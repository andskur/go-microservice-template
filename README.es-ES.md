

# go-microservice-template

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](https://go.dev/)

Una plantilla mínima de microservicios en Go con conexión CLI de Cobra/Viper, versionamiento impulsado por ldflags, registro de actividad con logrus, objetivos de Makefile, pruebas y CI/CD con GitHub Actions (lint/pruebas/construcción en PRs/main más lanzamientos auto-etiquetados en `main`). La estructura es intencionalmente simple para que puedas integrar tus cargas de trabajo de ejecución rápidamente.

## Inicio rápido
- Requisitos: Go 1.21+ (el módulo establece 1.23/1.24), GNU `make`.
- Clona y crea tu rama: `git checkout -b feature/your-branch`.
- Compilar: `make build` (binario `./microservice-template`).
- Ejecutar: `make run` (invoca `go run -race cmd/microservice-template.go serve`).
- Versión: `./microservice-template --version`.
- Lint: `make lint` (golangci-lint).
- Probar: `make test` o una prueba única `go test ./... -run TestName -count=1`.
- Generar + probar: `make test-with-gen` (ejecuta la generación de proto y swagger primero).
- Generar + lint: `make lint-with-gen` (ejecuta la generación de proto y swagger primero).
- Pruebas gRPC: `make test-grpc` (ejecuta el paquete gRPC incluyendo integración).
- Pruebas HTTP: `make test-http` (ejecuta las pruebas del paquete HTTP).
- Cobertura: `make test-coverage` (escribe `coverage.out`).
- Deps ordenadas: `make tidy`; actualizar deps: `make update`.
- Inicio rápido HTTP: consulta [docs/HTTP_SWAGGER_GUIDE.md](./docs/HTTP_SWAGGER_GUIDE.md); habilita con `HTTP_ENABLED=true`, genera la API con `make generate-api`, prueba con curl.
- Inicio rápido gRPC: consulta [docs/GRPC_GUIDE.md](./docs/GRPC_GUIDE.md); habilita con `GRPC_ENABLED=true`, prueba con grpcurl; usa protocolos compartidos desde `https://github.com/andskur/protocols-template.git`.

### Renombrar el proyecto
- Comando: `make rename NEW_NAME=my-service` (parámetro requerido).
- NUEVO_NOMBRE válido: letras minúsculas, números, guiones, segmentos opcionales con `/` (ej., `my-service`, `github.com/yourorg/my-service`).
- Actualizaciones: ruta del módulo e importaciones, vars del Makefile, archivo de entrada, `Use` raíz de Cobra, nombre de la estructura de la API swagger, binario del Dockerfile, referencias en README/AGENTS, remote opcional de git.
- Después del renombrado, regenera el código generado: `make generate-all` (o al menos `make generate-api` + generación de proto según sea necesario).
- Verifica después del renombrado: `go test ./...`, `make build`, `./<nuevo-binario> --version`.

## Características
- Simple, pequeña huella usando libs estándar más Cobra/Viper/logrus.
- **Sistema de módulos** para componentes opcionales (repositorio, servicio, HTTP, gRPC, cola, etc.).
- **API REST HTTP** con enfoque de especificación Swagger/OpenAPI primero usando go-swagger.
- **API gRPC** con generación de código de buffers de protocolo y chequeos de salud.
- Metadatos de versión inyectados vía ldflags (`pkg/version`).
- Registro estructurado vía singleton `pkg/logger`.
- Objetivos Makefile para build/run/lint/test/tidy/update/generate-api.
- Pipeline CI: lint/test/build en PRs y `main`; pipeline de lanzamiento auto-etiqueta en `main` y publica un lanzamiento en GitHub (solo código fuente).
- Pruebas incluidas para conexión CLI, valores predeterminados de config, versionamiento, singleton logger, helpers, módulos HTTP y gRPC.
- Amigable para renombrar: un solo nombre de marcador de posición con objetivo automatizado `make rename`.

## Estructura del proyecto
```
go-microservice-template/
├── api/                        # Especificaciones Swagger/OpenAPI
├── cmd/                        # Entrada CLI + comandos
├── config/                     # Predeterminados de Viper y esquema
├── db/migrations/              # Archivos de migración de base de datos (golang-migrate)
├── docs/                       # Guías adicionales (HTTP_SWAGGER_GUIDE, GRPC_GUIDE)
├── internal/
│   ├── application.go          # Conexión de App + registro de módulos
│   ├── grpc/                   # Módulo gRPC (servidor, interceptores)
│   ├── http/                   # Módulo HTTP (manejadores, middleware, autenticación)
│   ├── module/                 # Interfaz/gestor de módulos
│   ├── repository/             # Módulo de repositorio (opcional)
│   ├── service/                # Módulo de lógica de negocio
│   └── models/                 # Modelos/enums de dominio
├── pkg/                        # Paquetes reutilizables (logger, version)
├── protocols/                  # Definiciones de protocolo extraídas vía subtree (sin ejemplo integrado)
├── scripts/                    # Scripts de automatización (renombrar)
├── .github/workflows/          # Pipelines CI/CD
├── Dockerfile                  # Construcción de contenedor multi-etapa
├── docker-compose.yml          # Stack local (Postgres/app)
├── Makefile                    # Objetivos build/run/lint/test/proto/swagger
├── README.md, AGENTS.md        # Documentación y pautas
└── go.mod, go.sum              # Dependencias
```

## Sistema de módulos

Esta plantilla utiliza una **arquitectura basada en módulos** para componentes opcionales. Los módulos proporcionan un ciclo de vida estándar (Init → Start → Stop) y pueden habilitarse/deshabilitarse mediante configuración.

### Módulos disponibles

La plantilla incluye marcadores de posición de configuración para módulos comunes:

| Módulo | Propósito | Clave de Configuración | Estado |
|--------|---------|-----------|--------|
| Repositorio | Persistencia respaldada por base de datos (envuelve conexión DB) | `database` | ✅ Implementado (habilitado cuando `database.enabled` es true) |
| Servicio | Orquestador de lógica de negocio (deps opcionales) | n/a | ✅ Implementado (siempre registrado; repositorio opcional) |
| HTTP | Servidor API REST HTTP con Swagger/OpenAPI | `http` | ✅ Implementado (habilitado cuando `http.enabled` es true) |
| Servidor gRPC | Servidor API gRPC | `grpc` | ✅ Implementado (habilitado cuando `grpc.enabled` es true) |
| Cliente gRPC | Cliente de servicio externo para comunicación entre microservicios | `grpc_client` | ✅ Implementado (habilitado cuando `grpc_client.enabled` es true) |
| WebSocket | Servidor WebSocket con pub/sub y gestión de salas | `websocket` | ✅ Implementado (habilitado cuando `websocket.enabled` es true) |

### Habilitar módulos

La configuración puede provenir de variables de entorno (recomendado) o un archivo de configuración (`config.yaml` es opcional). Viper fusiona: flags > env vars > archivo de configuración.

**Ejemplo de env (preferido):**
```bash
export DATABASE_ENABLED=true
export DATABASE_DRIVER=postgres
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
```

**Ejemplo de config.yaml (opcional):**
```yaml
database:
  enabled: true
  driver: postgres
  host: localhost
  port: 5432
  # ... otros ajustes
```

- El módulo de repositorio se registra solo cuando `database.enabled` (o `DATABASE_ENABLED`) es `true`.
- El módulo de servicio siempre se registra; si no hay repositorio disponible, las operaciones de base de datos devuelven errores claros.

Consulta `config/scheme.go` para las definiciones de la estructura de configuración.

### Configuración de la base de datos

El módulo de repositorio requiere PostgreSQL cuando está habilitado.

**Inicio rápido con Docker:**
```bash
docker run --name postgres-dev \
  -e POSTGRES_USER=dev \
  -e POSTGRES_PASSWORD=dev \
  -e POSTGRES_DB=microservice_dev \
  -p 5432:5432 \
  -d postgres:16-alpine
```

**Instalar herramienta de migración (una vez):**
```bash
make migrate-install
```

**Ejecutar migraciones:**
```bash
# Aplicar todas las migraciones pendientes
make migrate-up

# Verificar versión actual de migración
make migrate-version
```

**Objetivos de migración disponibles:**
```bash
make migrate-install      # Instalar CLI golang-migrate
make migrate-create       # Crear nueva migración (requiere NAME=)
make migrate-up           # Aplicar todas las migraciones pendientes
make migrate-down         # Deshacer última migración
make migrate-force        # Forzar versión de migración (requiere VERSION=)
make migrate-version      # Mostrar versión actual de migración
make migrate-drop         # Eliminar todas las tablas (⚠️ PELIGRO - requiere confirmación)
```

**Desarrollo local con Docker Compose:**
```bash
# Iniciar Postgres y ejecutar migraciones automáticamente (usa db/migrations)
make compose-up

# Detener servicios
make compose-down

# Reiniciar servicios
make compose-restart
```

Para implementaciones en producción, ejecuta las migraciones antes de iniciar la aplicación o usa un trabajo de migración separado en tu pipeline de despliegue.

### Configuración de la API REST HTTP

El módulo HTTP proporciona una API REST con soporte de especificación Swagger/OpenAPI usando go-swagger.

**Instalar go-swagger (una vez):**
```bash
make swagger-install
```

**Generar código del servidor API desde la especificación:**
```bash
# Validar la especificación swagger
make swagger-validate

# Generar código del servidor desde api/swagger.yaml
make generate-api
```

**Habilitar módulo HTTP:**
```bash
export HTTP_ENABLED=true
export HTTP_HOST=0.0.0.0
export HTTP_PORT=8080
export HTTP_MOCK_AUTH=true  # Para desarrollo local (omite validación JWT)

# Ejecutar el servicio
make run
```

**Probar endpoints HTTP:**
```bash
# Chequeo de salud (endpoint público)
curl http://localhost:8080/health

# Obtener usuario por email (requiere auth en producción, modo mock para dev)
curl -H "Authorization: Bearer test-token" \
  "http://localhost:8080/users?email=test@example.com"
```

**Objetivos HTTP disponibles:**
```bash
make swagger-install      # Instalar CLI go-swagger
make swagger-validate     # Validar especificación swagger
make generate-api         # Generar código del servidor desde api/swagger.yaml
make swagger-clean        # Eliminar código generado
make test-http           # Ejecutar pruebas del módulo HTTP
```

**Opciones de configuración:**
- `http.enabled` - Habilitar/deshabilitar servidor HTTP (predeterminado: false)
- `http.host` - Host del servidor (predeterminado: 0.0.0.0)
- `http.port` - Puerto del servidor (predeterminado: 8080)
- `http.mock_auth` - Usar autenticación mock para desarrollo (predeterminado: false)
- `http.cors.enabled` - Habilitar CORS (predeterminado: true)
- `http.rate_limit.enabled` - Habilitar limitación de tasa (predeterminado: false)
- `http.rate_limit.requests_per_sec` - Límite de tasa (predeterminado: 100.0)

Para una guía detallada de desarrollo HTTP que incluye agregar nuevos endpoints, autenticación y middleware, consulta [docs/HTTP_SWAGGER_GUIDE.md](./docs/HTTP_SWAGGER_GUIDE.md).

### Configuración del cliente gRPC

El módulo de cliente gRPC habilita la comunicación con microservicios gRPC externos. Los manejadores HTTP usan el cliente para obtener datos de servicios externos, haciendo esta plantilla ideal para patrones de gateway/BFF (Backend para Frontend).

**Resumen de arquitectura:**
- **Enfoque híbrido**: Cliente gRPC puro en `pkg/userservice/` (solo tipos proto) + envoltorio de módulo en `internal/grpcclient/` (conversiones + ciclo de vida)
- **Integración HTTP**: Los manejadores HTTP reciben grpcClient como dependencia y lo usan para obtener datos de servicios externos
- **Independencia del servicio**: La capa de servicio permanece enfocada únicamente en la lógica de negocio local

**Habilitar cliente gRPC:**
```bash
export GRPC_CLIENT_ENABLED=true
export GRPC_CLIENT_ADDRESS="user-service:9090"
export GRPC_CLIENT_TIMEOUT="30s"

# Ejecutar el servicio
make run
```

**Opciones de configuración:**
- `grpc_client.enabled` - Habilitar/deshabilitar cliente (predeterminado: false)
- `grpc_client.address` - Dirección del servicio externo (predeterminado: "localhost:9090")
- `grpc_client.timeout` - Timeout de solicitud (predeterminado: "30s")
- `grpc_client.keep_alive.time` - Intervalo de ping keep-alive (predeterminado: "10s")
- `grpc_client.keep_alive.timeout` - Timeout keep-alive (predeterminado: "1s")
- `grpc_client.keep_alive.permit_without_stream` - Enviar pings sin streams (predeterminado: true)

**Patrones de integración de manejadores:**

La plantilla demuestra la obtención desde el servicio externo únicamente. Los patrones alternativos están documentados en el código del manejador:

```go
// Patrón 1: Solo externo (implementación actual)
user, err := h.grpcClient.GetUserByEmail(ctx, email)

// Patrón 2: Solo base de datos local (alternativa comentada)
// user, err := h.service.GetUserByEmail(ctx, email)

// Patrón 3: Externo con respaldo local (alternativa comentada)
// user, err := h.grpcClient.GetUserByEmail(ctx, email)
// if err != nil {
//     user, err = h.service.GetUserByEmail(ctx, email)
// }

// Patrón 4: Agregado de ambas fuentes (alternativa comentada)
// externalUser, _ := h.grpcClient.GetUserByEmail(ctx, email)
// localUser, _ := h.service.GetUserByEmail(ctx, email)
// user = mergeUsers(externalUser, localUser)
```

**Probar endpoints:**
```bash
# Requiere user-service externo ejecutándose en la dirección configurada
export HTTP_ENABLED=true
export HTTP_MOCK_AUTH=true
export GRPC_CLIENT_ENABLED=true
export GRPC_CLIENT_ADDRESS="user-service:9090"

make run

# Probar el endpoint
curl -H "Authorization: Bearer test-token" \
  "http://localhost:8080/users?email=test@example.com"
```

**Manejo de errores:**
- Devuelve **503 Service Unavailable** cuando grpcClient no está configurado
- Mapea errores gRPC: `not found` → 404, `invalid input` → 400, `unavailable` → 503

Para una guía detallada de desarrollo gRPC que incluye agregar nuevos servicios y definiciones proto, consulta [docs/GRPC_GUIDE.md](./docs/GRPC_GUIDE.md).

### Configuración del servidor WebSocket

El módulo WebSocket proporciona comunicación bidireccional en tiempo real con soporte pub/sub y gestión de salas usando `gorilla/websocket`.

**Habilitar módulo WebSocket:**
```bash
export WEBSOCKET_ENABLED=true
export WEBSOCKET_HOST=0.0.0.0
export WEBSOCKET_PORT=8081

# Ejecutar el servicio
make run
```

**Opciones de configuración:**
- `websocket.enabled` - Habilitar/deshabilitar servidor WebSocket (predeterminado: false)
- `websocket.host` - Host del servidor (predeterminado: 0.0.0.0)
- `websocket.port` - Puerto del servidor (predeterminado: 8081)
- `websocket.max_message_size` - Tamaño máximo de mensaje en bytes (predeterminado: 512000)
- `websocket.limits.max_connections` - Límite global de conexiones (predeterminado: 0 = ilimitado)
- `websocket.limits.max_connections_per_room` - Límite por sala (predeterminado: 0 = ilimitado)

**Conectar y probar con wscat:**
```bash
# Instalar wscat
npm install -g wscat

# Conectar
wscat -c ws://localhost:8081/ws

# Suscribirse a una sala
> {"type":"subscribe","room":"notifications"}
< {"type":"subscribed","room":"notifications","data":{"room":"notifications","client_count":1}}

# Publicar en una sala
> {"type":"publish","room":"notifications","data":{"message":"hello"}}

# Difundir a todos los clientes
> {"type":"broadcast","data":{"announcement":"server restart"}}
```

**Tipos de mensaje:**
- `subscribe` - Unirse a una sala
- `unsubscribe` - Salir de una sala
- `publish` - Enviar mensaje a una sala (debes estar suscrito)
- `broadcast` - Enviar mensaje a todos los clientes conectados
- `ping` - Ping de mantenimiento (el servidor responde con `pong`)

**Endpoint de salud:**
```bash
curl http://localhost:8081/health
# Devuelve: {"status":"healthy","clients":5,"rooms":2}
```

Para una guía detallada de desarrollo WebSocket, consulta [docs/WEBSOCKET_GUIDE.md](./docs/WEBSOCKET_GUIDE.md).

### Agregar módulos personalizados

Consulta [Guía de Desarrollo de Módulos](./docs/MODULE_DEVELOPMENT.md) para crear módulos personalizados. El sistema de módulos proporciona:

- **Ciclo de vida estándar**: Init → Start → Stop con chequeos de salud
- **Inyección de dependencias**: Los módulos dependen entre sí mediante inyección en el constructor (explícita; sin localizador de servicios)
- **Impulsado por configuración**: Habilitar/deshabilitar módulos vía YAML/env vars (el repositorio depende de `database.enabled`; el servicio siempre se registra)
- **Apagado graceful**: Limpieza automática en orden inverso de registro

## Modelos y Enumeraciones
- Ubicación: `internal/models` con etiquetas/ganchos go-pg para integración con base de datos.
- Validación: implementa `Validate() error` y devuelve `*models.ValidationError` (`Field`, `Message`) para errores estructurados.
- Enums: ints tipados con `String()` y `UserStatusFromString()` insensible a mayúsculas/minúsculas; agrega conversiones proto/JWT más tarde si es necesario.
- Ganchos: `BeforeInsert`/`BeforeUpdate` convierten enums a strings y aseguran UUID/timestamps; `AfterSelect` convierte strings de vuelta a enums.

### Ejemplo: Modelo Usuario
```go
// internal/models/user.go
user := &models.User{
    Email:  "test@example.com",
    Name:   "Jane Doe",
    Status: models.UserActive,
}

if err := user.Validate(); err != nil {
    if verr, ok := err.(*models.ValidationError); ok {
        // error estructurado con contexto de campo
        log.Printf("field=%s msg=%s", verr.Field, verr.Message)
    }
    return err
}
```

### Crear un nuevo modelo (patrón)
```go
// internal/models/widget.go
package models

type Widget struct {
    ID    uuid.UUID
    Name  string
    State WidgetState
}

func (w *Widget) Validate() error {
    if w.Name == "" {
        return newValidationError("name", "is required")
    }
    if w.State < WidgetActive || w.State >= widgetStateUnsupported {
        return newValidationError("state", "invalid value")
    }
    return nil
}
```

### Ejemplo: Enum UserStatus
```go
// internal/models/user_status.go
status := models.UserActive
fmt.Println(status.String()) // "active"

parsed, err := models.UserStatusFromString("DELETED")
if err != nil {
    // valor inválido
}
fmt.Println(parsed == models.UserDeleted) // true
```

### Crear un nuevo enum (patrón)
```go
// internal/models/widget_state.go
package models

type WidgetState int

const (
    WidgetActive WidgetState = iota
    WidgetDisabled
    widgetStateUnsupported
)

var widgetStates = [...]string{
    WidgetActive:   "active",
    WidgetDisabled: "disabled",
}

func (s WidgetState) String() string {
    if s < 0 || int(s) >= len(widgetStates) {
        return ""
    }
    return widgetStates[s]
}

func WidgetStateFromString(v string) (WidgetState, error) {
    for i, r := range widgetStates {
        if strings.EqualFold(v, r) {
            return WidgetState(i), nil
        }
    }
    return widgetStateUnsupported, fmt.Errorf("invalid widget state %q", v)
}
```

## Limitaciones
Esta es una plantilla básica y genérica de microservicios en Go diseñada para proporcionar una estructura clara y herramientas fundamentales. Permanece intencionalmente minimalista.

## Configuración
- Predeterminados: `env` predetermina a `prod` (`config/init.go:setDefaults`).
- Precedencia: flags > env vars > archivo de configuración.
- Nomenclatura de env vars: los puntos se convierten en underscores (reemplazador de Viper).
- Para agregar un campo de configuración:
  ```go
  // config/scheme.go
  type Scheme struct {
      Env  string // existente
      Port int    // nuevo
  }

  // config/init.go
  func setDefaults() {
      viper.SetDefault("env", "prod")
      viper.SetDefault("port", 8080)
  }

  // cmd/root (enlazar una flag)
  cmd.Flags().Int("port", 0, "port to listen on")
  ```
  La precedencia asegurará flag > env > archivo de configuración para `port` también.

## CLI
- Nombre del comando raíz: `microservice-template`.
- Subcomandos: `serve` (gancho de ejecución actual). Agrega más vía `cmd/<nombre>` y regístralos en la raíz.
- Salida de versión: `./microservice-template --version` (ldflags poblan `pkg/version`).
- Ciclo de vida `serve`: `PreRun` registra versión; `RunE` debe iniciar tus cargas de trabajo; `PostRun` siempre detiene la app.
- Agregar un nuevo comando (ejemplo):
  ```go
  // cmd/health/health.go
  package health

  import "github.com/spf13/cobra"

  func Cmd() *cobra.Command {
      return &cobra.Command{
          Use:   "health",
          Short: "Health probe",
          RunE: func(_ *cobra.Command, _ []string) error {
              // agregar verificaciones aquí
              return nil
          },
      }
  }
  ```
  Regístralo en `cmd/microservice-template.go`: `rootCmd.AddCommand(health.Cmd())`.

## Flujo de trabajo de desarrollo
- Formato: `gofmt` (usado vía herramientas de Go).
- Lint: `make lint` (golangci-lint; consulta `.golangci.yml`).
- Pruebas: `make test` o `go test ./...`; ejemplo de prueba única `go test ./cmd/root -run TestInitializeConfig -count=1`.
- Compilar: `make build` (CGO deshabilitado; ldflags inyectan info de versión).
- Deps: `make tidy` después de cambios; `make update` para actualizar módulos.

## CI/CD
- Flujos de trabajo: `.github/workflows/ci.yml` y `.github/workflows/release.yml`.
  - CI (`ci.yml`): en PRs y `main`, ejecuta `make lint`, `make test`, `make build` (Go 1.24) con caché de módulos.
  - Lanzamiento (`release.yml`): en `main`, vuelve a ejecutar lint/test/build, determina la siguiente etiqueta incremental (`v1`, `v2`, …), empuja la etiqueta y crea un lanzamiento en GitHub con notas autogeneradas (solo código fuente). Usa `GITHUB_TOKEN`; no se necesitan secretos adicionales.
- Protección de ramas (recomendado): exige que las verificaciones de CI (`lint`, `test`, `build`) pasen antes de fusionar en `main` y limita pushes directos.

## Versionamiento
- `Makefile` inyecta nombre/etiqueta/commit/rama/remote/fecha de compilación en `pkg/version` vía ldflags.
- `pkg/version` formatea una cadena de versión multi-línea y maneja valores no especificados.
- Ejemplo de salida:
  ```
  Template-service v0.0.0
  Branch main, commit hash: abcdef123
  Origin repository: https://github.com/org/repo
  Compiled at: 2026-01-16 20:58:09 +0000 UTC
  ©2026
  ```

## Registro de actividad (Logging)
- `pkg/logger.Log()` devuelve un logger de logrus con marcas de tiempo completas.
- Ejemplo:
  ```go
  log := logger.Log()
  log.Infof("starting service", "env=%s", cfg.Env)
  log.Errorf("failed to start: %v", err)
  ```

## Extender la plantilla
> **Nota:** Para renombrar un proyecto existente, consulta la sección [Renombrar el proyecto](#renaming-the-project) en Inicio rápido.

- Agregar config: actualiza `Scheme`, `setDefaults` y flags de CLI; prueba la binding como en `cmd/root/root_test.go`.
- Agregar comandos: crea `cmd/<nombre>` con `cobra.Command`, regístralo en la raíz en `cmd/microservice-template.go`.
  Después de renombrar el archivo de entrada (ej., `cmd/yourservice.go`), registra nuevos comandos allí.
- Agregar lógica de ejecución: implementa `App.Init/Serve/Stop` con manejo adecuado de contexto/apagado y apagado graceful.
- Agregar pruebas: sigue patrones tabulares; resetea estado global (Viper) en `t.Cleanup`.

## Mantenerse actualizado con los cambios de la plantilla

Este proyecto puede recibir actualizaciones desde la plantilla upstream: [go-microservice-template](https://github.com/andskur/go-microservice-template).

### Configuración inicial (proyectos downstream)
```bash
make template-setup
```
Esto:
- Agrega el remote de la plantilla (`template`)
- Obtiene los últimos cambios de la plantilla
- Crea `.template-version` para rastrear el estado de sincronización

### Verificar actualizaciones
```bash
make template-status
make template-diff       # resumen diff vs template/main
make template-diff v1.2.0 # diff contra una etiqueta específica
```

### Sincronizar actualizaciones
```bash
make template-fetch      # obtener últimos cambios de la plantilla
make template-sync       # fusionar template/main en la rama actual
make template-sync v1.2.0 # fusionar una etiqueta específica
```

Después de fusionar:
- Resuelve conflictos manualmente
- Ejecuta pruebas: `make test` (y `make build` si se desea)
- Commit con un mensaje claro (ej., `chore: sync from template v1.2.0`)

### Archivos que probablemente requieran atención durante la sincronización
- `README.md`, `AGENTS.md` (documentación específica del proyecto)
- `internal/application.go` (registro de módulos)
- `config/scheme.go` y `config/init.go` (esquema/predeterminados de configuración)
- `Makefile` (objetivos personalizados)

### Mejores prácticas
- Sincroniza regularmente para reducir conflictos
- Mantén los commits de sincronización separados del trabajo de características
- Revisa `make template-diff` antes de fusionar
- Usa `.template-version` para registrar la última referencia de plantilla sincronizada (se actualiza automáticamente en sincronizaciones exitosas)

## Contribuir
¡Las contribuciones son bienvenidas! No dudes en enviar un Pull Request.

Para pautas de desarrollo y mejores prácticas, consulta [AGENTS.md](./AGENTS.md).

## Licencia
Este proyecto está licenciado bajo la Licencia MIT — consulta el archivo [LICENSE](./LICENSE) para detalles.

## Autor
Copyright (c) 2022 Andrey Skurlatov
