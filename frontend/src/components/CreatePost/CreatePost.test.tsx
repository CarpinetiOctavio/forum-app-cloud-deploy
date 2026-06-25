import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreatePost } from './CreatePost';
import { postService } from '../../services/postService';

jest.mock('../../services/postService');
const mockedPostService = postService as jest.Mocked<typeof postService>;

describe('CreatePost Component', () => {
    const mockOnPostCreated = jest.fn();

    beforeEach(() => {
        jest.clearAllMocks();
    });

    test('renderiza el formulario correctamente', () => {
        render(<CreatePost userId={1} onPostCreated={mockOnPostCreated} />);

        expect(screen.getByText('Create New Post')).toBeInTheDocument();
        expect(screen.getByPlaceholderText(/write a title/i)).toBeInTheDocument();
        expect(screen.getByPlaceholderText(/what do you want to share/i)).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /publish post/i })).toBeInTheDocument();
    });

    test('crea post exitosamente', async () => {
        mockedPostService.createPost.mockResolvedValueOnce({
            id: 1,
            title: 'Test Post',
            content: 'Test Content',
            user_id: 1,
            username: 'testuser',
            created_at: '2024-01-01'
        });

        render(<CreatePost userId={1} onPostCreated={mockOnPostCreated} />);

        const titleInput = screen.getByPlaceholderText(/write a title/i);
        const contentInput = screen.getByPlaceholderText(/what do you want to share/i);
        const submitButton = screen.getByRole('button', { name: /publish post/i });

        fireEvent.change(titleInput, { target: { value: 'Test Post' } });
        fireEvent.change(contentInput, { target: { value: 'Test Content' } });
        fireEvent.click(submitButton);

        await waitFor(() => {
            expect(mockedPostService.createPost).toHaveBeenCalledWith(
                { title: 'Test Post', content: 'Test Content' },
                1
            );
            expect(mockOnPostCreated).toHaveBeenCalled();
        });
    });

    test('muestra error cuando falla la creación', async () => {
        mockedPostService.createPost.mockRejectedValueOnce({
            response: { data: { error: 'Error al crear post' } }
        });

        render(<CreatePost userId={1} onPostCreated={mockOnPostCreated} />);

        fireEvent.change(screen.getByPlaceholderText(/write a title/i), { target: { value: 'Test' } });
        fireEvent.change(screen.getByPlaceholderText(/what do you want to share/i), { target: { value: 'Content' } });
        fireEvent.click(screen.getByRole('button', { name: /publish post/i }));

        await waitFor(() => {
            expect(screen.getByText('Error al crear post')).toBeInTheDocument();
        });

        expect(mockOnPostCreated).not.toHaveBeenCalled();
    });

    test('deshabilita el botón mientras está cargando', async () => {
        mockedPostService.createPost.mockImplementation(
            () => new Promise(resolve => setTimeout(resolve, 100))
        );

        render(<CreatePost userId={1} onPostCreated={mockOnPostCreated} />);

        fireEvent.change(screen.getByPlaceholderText(/write a title/i), { target: { value: 'Test' } });
        fireEvent.change(screen.getByPlaceholderText(/what do you want to share/i), { target: { value: 'Content' } });

        const submitButton = screen.getByRole('button', { name: /publish post/i });
        fireEvent.click(submitButton);

        expect(submitButton).toBeDisabled();
        expect(screen.getByText('Publishing...')).toBeInTheDocument();
    });

    test('should show fallback error when response has no error detail', async () => {
        mockedPostService.createPost.mockRejectedValueOnce(new Error('Network error'));

        render(<CreatePost userId={1} onPostCreated={mockOnPostCreated} />);

        fireEvent.change(screen.getByPlaceholderText(/write a title/i), { target: { value: 'Test' } });
        fireEvent.change(screen.getByPlaceholderText(/what do you want to share/i), { target: { value: 'Content' } });
        fireEvent.click(screen.getByRole('button', { name: /publish post/i }));

        await waitFor(() => {
            expect(screen.getByText('Failed to create post')).toBeInTheDocument();
        });

        expect(mockOnPostCreated).not.toHaveBeenCalled();
    });
});