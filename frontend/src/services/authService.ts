import axios from 'axios';
import { User, LoginRequest, RegisterRequest } from '../types';

const API_BASE = process.env.REACT_APP_API_URL || 'http://localhost:8080';
const API_URL = `${API_BASE}/api/auth`;

export const authService = {
  // User login
  async login(credentials: LoginRequest): Promise<User> {
    const response = await axios.post<User>(`${API_URL}/login`, credentials);
    return response.data;
  },

  // User registration
  async register(data: RegisterRequest): Promise<User> {
    const response = await axios.post<User>(`${API_URL}/register`, data);
    return response.data;
  }
};