import axios from 'axios';
import { authService } from './authService';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('authService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('login', () => {
    test('llama a la API correctamente con credentials', async () => {
      const mockUser = {
        id: 1,
        email: 'test@example.com',
        username: 'testuser',
        created_at: '2025-01-01'
      };

      mockedAxios.post.mockResolvedValueOnce({ data: mockUser });

      const result = await authService.login({
        email: 'test@example.com',
        password: '123456'
      });

      expect(mockedAxios.post).toHaveBeenCalledWith(
        'http://localhost:8080/api/auth/login',
        {
          email: 'test@example.com',
          password: '123456'
        }
      );
      expect(result).toEqual(mockUser);
    });

    test('rechaza cuando las credenciales son inválidas', async () => {
      const error = new Error('Credenciales inválidas');
      mockedAxios.post.mockRejectedValueOnce(error);

      await expect(
        authService.login({
          email: 'wrong@example.com',
          password: 'wrong'
        })
      ).rejects.toEqual(error);
    });
  });

  describe('register', () => {
    test('llama a la API correctamente con datos de registro', async () => {
      const mockUser = {
        id: 1,
        email: 'newuser@example.com',
        username: 'newuser',
        created_at: '2025-01-01'
      };

      mockedAxios.post.mockResolvedValueOnce({ data: mockUser });

      const result = await authService.register({
        email: 'newuser@example.com',
        password: '123456',
        username: 'newuser'
      });

      expect(mockedAxios.post).toHaveBeenCalledWith(
        'http://localhost:8080/api/auth/register',
        {
          email: 'newuser@example.com',
          password: '123456',
          username: 'newuser'
        }
      );
      expect(result).toEqual(mockUser);
    });

    test('rechaza cuando el email ya existe', async () => {
      const error = new Error('el email ya está registrado');
      mockedAxios.post.mockRejectedValueOnce(error);

      await expect(
        authService.register({
          email: 'existing@example.com',
          password: '123456',
          username: 'testuser'
        })
      ).rejects.toEqual(error);
    });

    test('rechaza cuando la validación falla', async () => {
      const error = new Error('la contraseña debe tener al menos 6 caracteres');
      mockedAxios.post.mockRejectedValueOnce(error);

      await expect(
        authService.register({
          email: 'test@example.com',
          password: '123',
          username: 'testuser'
        })
      ).rejects.toEqual(error);
    });
  });
});