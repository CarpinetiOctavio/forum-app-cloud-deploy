import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { Login } from './Login';
import axios from 'axios';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('Login Component', () => {
  const mockOnLoginSuccess = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders the login form correctly', () => {
    render(<Login onLoginSuccess={mockOnLoginSuccess} />);

    // Verify the heading renders
    expect(screen.getByRole('heading', { name: /sign in/i })).toBeInTheDocument();
    
    // Verify inputs
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    
    // Verify submit button
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });

  test('shows the registration form when the mode is toggled', () => {
    render(<Login onLoginSuccess={mockOnLoginSuccess} />);

    // Click the button to switch to register mode
    const toggleButton = screen.getByText(/don't have an account/i);
    fireEvent.click(toggleButton);

    // Verify it shows the Sign Up heading
    expect(screen.getByRole('heading', { name: /sign up/i })).toBeInTheDocument();
    
    // Verify the username field appears
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument();
  });

  test('successful login calls onLoginSuccess', async () => {
    const mockUser = {
      id: 1,
      email: 'test@example.com',
      username: 'testuser',
      created_at: '2025-01-01'
    };

    mockedAxios.post.mockResolvedValueOnce({ data: mockUser });

    render(<Login onLoginSuccess={mockOnLoginSuccess} />);

    // Fill out the form
    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    const passwordInput = screen.getByLabelText(/password/i) as HTMLInputElement;

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: '123456' } });

    // Submit
    const submitButton = screen.getByRole('button', { name: /sign in/i });
    fireEvent.click(submitButton);

    // Verify
    await waitFor(() => {
      expect(mockOnLoginSuccess).toHaveBeenCalledWith(mockUser);
    });
  });

  test('successful registration calls authService.register and onLoginSuccess', async () => {
    const mockUser = {
      id: 2,
      email: 'new@example.com',
      username: 'newuser',
      created_at: '2025-01-01'
    };

    mockedAxios.post.mockResolvedValueOnce({ data: mockUser });

    render(<Login onLoginSuccess={mockOnLoginSuccess} />);

    // Switch to register mode
    fireEvent.click(screen.getByText(/don't have an account/i));

    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    const usernameInput = screen.getByLabelText(/username/i) as HTMLInputElement;
    const passwordInput = screen.getByLabelText(/password/i) as HTMLInputElement;

    fireEvent.change(emailInput, { target: { value: 'new@example.com' } });
    fireEvent.change(usernameInput, { target: { value: 'newuser' } });
    fireEvent.change(passwordInput, { target: { value: '123456' } });

    const submitButton = screen.getByRole('button', { name: /sign up/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockedAxios.post).toHaveBeenCalledWith(
        'http://localhost:8080/api/auth/register',
        {
          email: 'new@example.com',
          password: '123456',
          username: 'newuser'
        }
      );
      expect(mockOnLoginSuccess).toHaveBeenCalledWith(mockUser);
    });
  });

  test('shows an error message when login fails', async () => {
    mockedAxios.post.mockRejectedValueOnce({
      response: {
        data: {
          error: 'Invalid credentials'
        }
      }
    });

    render(<Login onLoginSuccess={mockOnLoginSuccess} />);

    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    const passwordInput = screen.getByLabelText(/password/i) as HTMLInputElement;

    fireEvent.change(emailInput, { target: { value: 'wrong@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'wrongpass' } });

    const submitButton = screen.getByRole('button', { name: /sign in/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('Invalid credentials')).toBeInTheDocument();
    });

    expect(mockOnLoginSuccess).not.toHaveBeenCalled();
  });

  test('disables the submit button while the request is loading', async () => {
    mockedAxios.post.mockImplementation(() => 
      new Promise(resolve => setTimeout(resolve, 100))
    );

    render(<Login onLoginSuccess={mockOnLoginSuccess} />);

    const emailInput = screen.getByLabelText(/email/i) as HTMLInputElement;
    const passwordInput = screen.getByLabelText(/password/i) as HTMLInputElement;

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: '123456' } });

    const submitButton = screen.getByRole('button', { name: /sign in/i });
    fireEvent.click(submitButton);

    // The button must be disabled
    expect(submitButton).toBeDisabled();
  });
});