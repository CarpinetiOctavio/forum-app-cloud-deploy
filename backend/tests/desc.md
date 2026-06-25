# Tests del Backend - Capa de Pruebas Unitarias

## ¿Qué es esta carpeta?

Contiene **pruebas unitarias** que verifican que la lógica de negocio funciona correctamente sin tocar la base de datos real.

## Estructura

```
tests/
├── mocks/                    # Objetos FALSOS (simulan la BD)
│   ├── user_repository_mock.go
│   └── post_repository_mock.go
└── services/                 # Tests de la lógica de negocio
    ├── auth_service_test.go
    └── post_service_test.go
```

## ¿Por qué está separado así?

### `/mocks`
Contiene implementaciones FALSAS de los repositorios.

**Ejemplo:**
```go
// Repository REAL (toca la BD)
func (r *SQLiteUserRepository) FindByEmail(email string) (*models.User, error) {
    query := `SELECT * FROM users WHERE email = ?`
    row := r.db.QueryRow(query, email)  // ← Toca la BD real
    ...
}

// Repository FALSO (no toca nada)
func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
    args := m.Called(email)  // ← Devuelve lo que configuremos
    return args.Get(0).(*models.User), args.Error(1)
}
```

### `/services`
Contiene tests que prueban la lógica de negocio usando los mocks.

**Ejemplo:**
```go
func TestRegister_Success(t *testing.T) {
    mockRepo := new(mocks.MockUserRepository)  // ← Mock, NO BD real
    mockRepo.On("FindByEmail", "test@example.com").Return(nil, nil)
    
    authService := services.NewAuthService(mockRepo)
    user, err := authService.Register(...)
    
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)  // ← Verificar que se llamó correctamente
}
```

## Patrón AAA (Arrange, Act, Assert)

Cada test sigue este patrón:

```go
// ARRANGE: Preparar datos y mocks
mockRepo := new(mocks.MockUserRepository)
mockRepo.On("FindByEmail", "test@example.com").Return(nil, nil)

// ACT: Ejecutar la función
user, err := authService.Register(&models.RegisterRequest{...})

// ASSERT: Verificar el resultado
assert.NoError(t, err)
assert.Equal(t, "test@example.com", user.Email)
```

## Tests cubiertos

### AuthService (11 tests)
- **Register:** email vacío, email inválido, password corto, username vacío, email duplicado, registro exitoso
- **Login:** login exitoso, email vacío, password vacío, usuario no existe, password incorrecta

### PostService (8 tests)
- **CreatePost:** crear exitosamente, usuario no existe, error del repositorio, título vacío, contenido vacío
- **DeletePost:** eliminar exitosamente, post no existe, usuario no es autor (REGLA DE NEGOCIO)

## ¿Por qué no toca la BD?

**Ventajas:**

1. **Rápido**: Los tests corren en milisegundos
2. **Aislado**: No depende de que la BD esté corriendo
3. **Repetible**: Siempre da los mismos resultados
4. **Independiente**: No contamina la BD de producción

**Comparación:**

```
SIN mocks (toca BD real):
- Lento (segundos)
- Depende de la BD (si cae, fallan los tests)
- Modifican datos (contamina la BD)
- No reproduces errores fácilmente

CON mocks:
- Rápido (milisegundos)
- Funciona sin BD
- No modifica nada
- Puedes reproducir cualquier escenario
```

## Cómo ejecutar

```bash
# Todos los tests
go test ./tests/services/... -v

# Con cobertura
go test ./tests/services/... -v -cover

# Solo un test específico
go test ./tests/services/ -v -run TestRegister_Success
```

## Concepto clave: AISLAMIENTO

El objetivo de los mocks es **aislar** lo que estás probando:

```
Testea ESTO:        NO testea ESTO:
┌──────────────┐   ┌──────────────┐
│ Validaciones │   │ Base de datos │ (mockeado)
│ Lógica       │   │ Archivos      │ (no necesarios)
│ Reglas       │   │ APIs externas │ (no necesarias)
└──────────────┘   └──────────────┘
```

## Relación con productivo

**En tests:**
```go
userRepo := new(mocks.MockUserRepository)  // ← MOCK
```

**En producción:**
```go
db := database.InitDB("./database.db")
userRepo := repository.NewSQLiteUserRepository(db)  // ← REAL
```

El código de negocio es el mismo, solo cambia la implementación del repositorio.