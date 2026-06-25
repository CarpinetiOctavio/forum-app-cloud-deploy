# Tests del Frontend - Pruebas de Componentes y Servicios

## ¿Qué es esta carpeta?

Contiene tests unitarios para componentes React y servicios HTTP. Verifica que la interfaz funcione correctamente sin tocar el servidor real.

## Estructura

```
src/
├── components/
│   ├── Login/
│   │   ├── Login.tsx           ← Componente REAL
│   │   └── Login.test.tsx      ← Tests del componente
│   └── PostList/
│       ├── PostList.tsx        ← Componente REAL
│       └── PostList.test.tsx   ← Tests del componente
├── services/
│   ├── authService.ts          ← Servicio REAL
│   └── authService.test.ts     ← Tests del servicio
└── __mocks__/
    └── axios.ts                ← Mock del HTTP
```

## .tsx vs .test.tsx

### Login.tsx (COMPONENTE REAL)

```typescript
// Código de PRODUCCIÓN
export const Login: React.FC<LoginProps> = ({ onLoginSuccess }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  
  const handleSubmit = async (e) => {
    const user = await authService.login({ email, password });
    onLoginSuccess(user);
  };
  
  return (
    <form onSubmit={handleSubmit}>
      <input value={email} onChange={...} />
      <button>Iniciar Sesión</button>
    </form>
  );
};
```

### Login.test.tsx (TESTS DEL COMPONENTE)

```typescript
// Código de TESTING
describe('Login Component', () => {
  test('renderiza el formulario correctamente', () => {
    render(<Login onLoginSuccess={mockFn} />);
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
  });
  
  test('login exitoso llama a onLoginSuccess', async () => {
    mockedAxios.post.mockResolvedValueOnce({ data: mockUser });
    render(<Login onLoginSuccess={mockFn} />);
    fireEvent.click(screen.getByRole('button', { name: /iniciar/i }));
    await waitFor(() => {
      expect(mockFn).toHaveBeenCalledWith(mockUser);
    });
  });
});
```

## ¿Cuál es la diferencia?

| Aspecto | .tsx | .test.tsx |
|---------|------|-----------|
| Código | PRODUCCIÓN (usuarios lo ven) | TESTING (desarrollo) |
| Ejecución | `npm start` (app corriendo) | `npm test` (tests) |
| Propósito | Funcionalidad real | Verificar que funciona |
| Depuración | Errores en navegador | Errores en terminal |

## Mocking de Axios

### Problema sin mock

```typescript
// Sin mock (MALO)
test('login', async () => {
  const user = await authService.login({ email, password });
  // ✗ Hace una petición HTTP REAL a localhost:8080
  // ✗ Necesita que el backend esté corriendo
  // ✗ Es lento
  // ✗ Puede fallar por razones externas
});
```

### Solución con mock

```typescript
// Con mock (BIEN)
jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

test('login', async () => {
  mockedAxios.post.mockResolvedValueOnce({ data: mockUser });
  // ✓ No hace petición HTTP real
  // ✓ No necesita el backend
  // ✓ Es rápido
  // ✓ Siempre produce el mismo resultado
});
```

## Patrón AAA en Frontend

```typescript
describe('Login Component', () => {
  test('login exitoso', async () => {
    // ARRANGE: Preparar mocks y renderizar
    mockedAxios.post.mockResolvedValueOnce({ data: mockUser });
    const mockOnLoginSuccess = jest.fn();
    render(<Login onLoginSuccess={mockOnLoginSuccess} />);
    
    // ACT: Ejecutar acciones del usuario
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: 'test@example.com' }
    });
    fireEvent.click(screen.getByRole('button', { name: /iniciar/i }));
    
    // ASSERT: Verificar el resultado
    await waitFor(() => {
      expect(mockOnLoginSuccess).toHaveBeenCalledWith(mockUser);
    });
  });
});
```

## Tests cubiertos

### Login Component (5 tests)
- Renderiza formulario correctamente
- Cambia entre login y registro
- Login exitoso llama callback
- Muestra error cuando falla
- Deshabilita botón mientras carga

### PostList Component (5 tests)
- Renderiza lista de posts
- Muestra "No hay posts" cuando está vacía
- Muestra botón eliminar solo para posts propios
- Elimina post cuando se hace click
- Muestra error cuando falla cargar

### authService (4 tests)
- Login exitoso
- Login fallido
- Register exitoso
- Register falla si email existe

## ¿Por qué no tocas el servidor?

**Ventajas de mockear HTTP:**

1. **Rápido**: No espera respuesta del servidor
2. **Confiable**: Siempre da el mismo resultado
3. **Independiente**: Funciona sin backend corriendo
4. **Predecible**: Puedes simular cualquier escenario

## Concepto clave: Testing de UI

Probás:
- ✓ ¿Se renderiza correctamente?
- ✓ ¿Responde a clicks?
- ✓ ¿Valida inputs?
- ✓ ¿Llama las funciones esperadas?

NO probás:
- ✗ El backend (eso tienen sus tests)
- ✗ La BD (el backend lo hace)
- ✗ HTTP real (es responsabilidad del backend)

## Cómo ejecutar

```bash
# Todos los tests
npm test

# Con cobertura
npm test -- --coverage

# Un archivo específico
npm test Login.test.tsx

# En modo watch (se ejecutan automáticamente)
npm test -- --watch
```

## Relación con backend

```
Backend tests:
  - Prueba AuthService.Login (con mock repo)

Frontend tests:
  - Prueba Login Component (con mock axios)
  - Pero axios llamaría al backend en producción
```

Ambos testan lo suyo, en paralelo, sin depender el uno del otro.