# Services - Capa de Comunicación HTTP

## ¿Qué hace esta carpeta?

Contiene funciones que comunican con el backend usando HTTP. Son el "traductor" entre componentes React y la API del servidor.

## Estructura

```
services/
├── authService.ts       ← Llamadas HTTP de autenticación
├── authService.test.ts  ← Tests del servicio
├── postService.ts       ← Llamadas HTTP de posts
└── postService.test.ts  ← Tests del servicio
```

## ¿Por qué separar en services?

### Problema (sin services)

```typescript
// ❌ MALO: HTTP mezclado en componentes
const Login = () => {
  const handleLogin = async () => {
    const response = await axios.post(
      'http://localhost:8080/api/auth/login',
      { email, password }
    );
    setUser(response.data);
  };
};

// Problemas:
// - URL duplicada en varios componentes
// - Lógica HTTP esparcida
// - Difícil de testear
// - Difícil de cambiar (si cambias la URL, editas todo)
```

### Solución (con services)

```typescript
// ✓ BIEN: HTTP centralizado
const Login = () => {
  const handleLogin = async () => {
    const user = await authService.login({ email, password });
    setUser(user);
  };
};

// authService.ts
export const authService = {
  async login(credentials) {
    const response = await axios.post(`${API_URL}/login`, credentials);
    return response.data;
  }
};
```

## Beneficios

1. **Centralizado**: Toda la lógica HTTP en un lugar
2. **Reutilizable**: Múltiples componentes usan el mismo service
3. **Testeable**: Fácil de mockear axios
4. **Mantenible**: Cambias un lugar, no diez

## Cómo funciona

### authService.ts

```typescript
const API_URL = 'http://localhost:8080/api/auth';

export const authService = {
  async login(credentials: LoginRequest): Promise<User> {
    const response = await axios.post<User>(
      `${API_URL}/login`,
      credentials
    );
    return response.data;  // ← Devuelve solo los datos, no la respuesta completa
  },

  async register(data: RegisterRequest): Promise<User> {
    const response = await axios.post<User>(
      `${API_URL}/register`,
      data
    );
    return response.data;
  }
};
```

### Cómo lo usa un componente

```typescript
const Login = () => {
  const handleSubmit = async () => {
    try {
      const user = await authService.login({ email, password });
      onLoginSuccess(user);  // El componente recibe el usuario
    } catch (error) {
      setError(error.message);  // Y maneja el error
    }
  };
};
```

## postService.ts

Similar pero para posts:

```typescript
export const postService = {
  async getAllPosts(): Promise<Post[]> {
    const response = await axios.get<Post[]>(API_URL);
    return response.data;
  },

  async createPost(data: CreatePostRequest, userId: number): Promise<Post> {
    const response = await axios.post<Post>(API_URL, data, {
      headers: { 'X-User-ID': userId.toString() }
    });
    return response.data;
  },

  async deletePost(id: number, userId: number): Promise<void> {
    await axios.delete(`${API_URL}/${id}`, {
      headers: { 'X-User-ID': userId.toString() }
    });
  }
};
```

## Testing de Services

### Cómo lo testeas (CON MOCK)

```typescript
jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

test('login exitoso', async () => {
  const mockUser = { id: 1, email: 'test@example.com', ... };
  mockedAxios.post.mockResolvedValueOnce({ data: mockUser });
  
  const result = await authService.login({
    email: 'test@example.com',
    password: '123456'
  });
  
  expect(result).toEqual(mockUser);
});
```

### Por qué mockeas axios

```
SIN mock:
  authService.login()
    ↓
  axios.post() ← Va al servidor real
    ↓
  Espera respuesta (lento)
  
CON mock:
  authService.login()
    ↓
  axios.post() ← MOCKEADO, devuelve al instante
    ↓
  Resultado inmediato (rápido)
```

## Concepto clave: Abstracción

El componente **NO necesita saber** que estás usando axios:

```typescript
// El componente solo sabe:
const user = await authService.login(credentials);

// No sabe (y no le importa):
// - Que estamos usando axios
// - Que hay HTTP involucrado
// - Cuál es la URL exacta
```

Si mañana cambias de axios a fetch o a otra librería, el componente no se entera.

## Relación con Backend

```
Frontend (React)          Backend (Go)
┌─────────────┐          ┌──────────────┐
│ authService │─────────→│ AuthHandler  │
│  (HTTP)     │  HTTP    │  (recepta)   │
└─────────────┘          └──────────────┘
     ↓                         ↓
Mock en tests         Mock en tests
(no toca servidor)    (no toca BD)
```

## Resumen

- **authService.ts**: Funciones para login/register
- **postService.ts**: Funciones para posts/comentarios
- **Tests**: Mockean axios para evitar HTTP real
- **Propósito**: Centralizar toda la lógica HTTP