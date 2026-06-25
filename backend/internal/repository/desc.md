# Repository - Capa de Acceso a Datos

## ¿Qué hace esta capa?

La capa **Repository** se encarga de **toda la interacción con la base de datos**. Es la única capa que ejecuta queries SQL.

## Patrón Repository

Usamos el patrón Repository para:
1. **Aislar** la lógica de acceso a datos del resto de la aplicación
2. **Facilitar el testing** mediante el uso de interfaces
3. **Centralizar** todas las operaciones de base de datos en un solo lugar

## Estructura

### Interfaces (Contratos)
- `UserRepository`: Define operaciones sobre usuarios
- `PostRepository`: Define operaciones sobre posts y comentarios

### Implementaciones
- `SQLiteUserRepository`: Implementación real usando SQLite
- `SQLitePostRepository`: Implementación real usando SQLite

## ¿Por qué interfaces?

Las **interfaces** permiten:
- En **producción**: usar la implementación real (`SQLiteUserRepository`)
- En **tests**: usar **mocks** (implementaciones falsas) para NO tocar la base de datos real

Ejemplo:
```go
// Interface (contrato)
type UserRepository interface {
    Create(user *models.User) error
    FindByEmail(email string) (*models.User, error)
}

// En producción
var repo UserRepository = NewSQLiteUserRepository(db)

// En tests
var repo UserRepository = NewMockUserRepository() // Mock
```

## Operaciones disponibles

### UserRepository
- `Create()`: Crea un nuevo usuario
- `FindByEmail()`: Busca usuario por email (para login)
- `FindByID()`: Busca usuario por ID

### PostRepository
- `Create()`: Crea un nuevo post
- `FindAll()`: Obtiene todos los posts
- `FindByID()`: Busca un post específico
- `Delete()`: Elimina un post
- `CreateComment()`: Agrega un comentario a un post
- `FindCommentsByPostID()`: Obtiene comentarios de un post

## Principio de responsabilidad única

Esta capa **SOLO** se encarga de:
- ✅ Ejecutar queries SQL
- ✅ Mapear resultados a structs

Esta capa **NO** se encarga de:
- ❌ Validar datos (eso es del Service)
- ❌ Manejar HTTP (eso es del Handler)
- ❌ Lógica de negocio (eso es del Service)