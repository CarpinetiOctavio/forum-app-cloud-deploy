# Database - Inicialización de la Base de Datos

## ¿Qué hace esta carpeta?

Se encarga de **crear e inicializar la base de datos SQLite** con el schema (estructura) necesario.

## ¿Por qué SQLite?

```
SQLite:
✓ No necesita servidor separado
✓ Es solo un archivo (.db)
✓ Perfecto para desarrollo y testing
✓ Se crea automáticamente

PostgreSQL/MySQL:
✗ Necesita servidor corriendo
✗ Más complejo para desarrollo local
✗ Mejor para producción
```

## Cómo funciona

### 1. Cuando levantás la app

```bash
go run cmd/api/main.go
```

**Qué sucede:**
```go
db, err := database.InitDB("./database.db")
// 1. Busca el archivo database.db
// 2. Si no existe, lo crea
// 3. Si existe, abre la conexión existente
// 4. Ejecuta las sentencias CREATE TABLE IF NOT EXISTS
// 5. Crea índices
```

### 2. Resultado

```
tp06-testing/
├── backend/
│   ├── database.db          ← Archivo creado automáticamente
│   ├── cmd/api/
│   │   └── main.go
│   └── internal/
│       └── database/
│           └── database.go
```

El archivo `database.db` contiene:
- Tabla `users`
- Tabla `posts`
- Tabla `comments`
- Índices para optimizar búsquedas

## Schema (Estructura)

```sql
users
├── id (PRIMARY KEY)
├── email (UNIQUE)
├── password
├── username
└── created_at

posts
├── id (PRIMARY KEY)
├── title
├── content
├── user_id (FOREIGN KEY → users)
└── created_at

comments
├── id (PRIMARY KEY)
├── post_id (FOREIGN KEY → posts)
├── user_id (FOREIGN KEY → users)
├── content
└── created_at
```

## Relaciones

```
users (1) ──→ (*) posts
                    ↓
              (1) ──→ (*) comments
                      ↑
              (1) ──←── users
```

**Ejemplo:**
- Usuario 1 crea 3 posts
- Usuarios 2 y 3 comentan en esos posts
- Si borras el usuario 1, se borran sus posts y sus comentarios

## ¿Por qué "IF NOT EXISTS"?

```sql
CREATE TABLE IF NOT EXISTS users (...)
```

Esto permite:
1. Ejecutar InitDB múltiples veces sin errores
2. Desarrollar sin borrar datos anterior
3. Migrations futuras más fáciles

## En Tests

**Los tests NO usan el archivo database.db**

```go
// En main.go (producción)
db, _ := database.InitDB("./database.db")  // Toca el archivo real

// En tests
mockRepo := new(mocks.MockUserRepository)  // No toca nada
```

**Por eso los tests son rápidos e independientes**

## Relación con Repository

```
database.go       ← Define el schema
    ↓
repository.go     ← Usa la BD para hacer queries
    ↓
services.go       ← Lógica de negocio (usa repository)
    ↓
handlers.go       ← HTTP (usa services)
```

## Cómo verificar que funciona

```bash
# 1. Ejecutar la app
go run cmd/api/main.go

# 2. En otra terminal, verificar que se creó
ls -la backend/database.db

# 3. Hacer un request para crear datos
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"123456","username":"testuser"}'

# 4. Inspeccionar la BD (si tienes sqlite3)
sqlite3 database.db "SELECT * FROM users;"
```

## Concepto clave: Separación de preocupaciones

Esta carpeta SOLO se encarga de:
- ✓ Definir el schema
- ✓ Crear tablas
- ✓ Abrir conexión

NO se encarga de:
- ✗ Ejecutar queries (eso es repository)
- ✗ Validar datos (eso es services)
- ✗ Manejar HTTP (eso es handlers)