# Services - Capa de Lógica de Negocio

## ¿Qué hace esta capa?

La capa **Services** contiene toda la **lógica de negocio** y **validaciones** de la aplicación.

## Responsabilidades

Esta capa se encarga de:
- ✅ Validar datos de entrada
- ✅ Aplicar reglas de negocio
- ✅ Coordinar operaciones entre repositorios
- ✅ Transformar datos si es necesario

Esta capa **NO** se encarga de:
- ❌ Manejar HTTP requests/responses (eso es del Handler)
- ❌ Ejecutar SQL directamente (eso es del Repository)

## Services disponibles

### AuthService
Maneja autenticación y registro de usuarios.

**Métodos:**
- `Register()`: Registra un nuevo usuario
  - Valida email (no vacío, contiene @)
  - Valida password (mínimo 6 caracteres)
  - Valida username (no vacío)
  - Verifica que el email no esté duplicado

- `Login()`: Autentica un usuario
  - Valida credenciales
  - Verifica que el usuario exista
  - Verifica que la contraseña coincida

### PostService
Maneja posts y comentarios.

**Métodos:**
- `CreatePost()`: Crea un nuevo post
  - Valida título (no vacío, mínimo 3 caracteres)
  - Valida contenido (no vacío)
  - Verifica que el usuario exista

- `GetAllPosts()`: Obtiene todos los posts

- `GetPostByID()`: Obtiene un post específico
  - Valida que el ID sea válido
  - Verifica que el post exista

- `DeletePost()`: Elimina un post
  - Verifica que el post exista
  - **Regla de negocio**: Solo el autor puede eliminar su post

- `CreateComment()`: Agrega un comentario
  - Valida contenido no vacío
  - Verifica que el post exista
  - Verifica que el usuario exista

- `GetCommentsByPostID()`: Obtiene comentarios de un post

## Inyección de dependencias

Los services reciben repositories a través de sus constructores:

```go
func NewAuthService(userRepo repository.UserRepository) *AuthService {
    return &AuthService{userRepo: userRepo}
}
```

Esto permite:
- En **producción**: inyectar repositorios reales
- En **tests**: inyectar mocks

## Casos de prueba importantes

Esta capa es la **MÁS IMPORTANTE para testing** porque contiene todas las validaciones.

Ejemplos de tests que haremos:
- ✅ Registrar con email vacío → debe fallar
- ✅ Crear post con título corto → debe fallar
- ✅ Usuario intenta eliminar post de otro → debe fallar
- ✅ Comentar en post inexistente → debe fallar