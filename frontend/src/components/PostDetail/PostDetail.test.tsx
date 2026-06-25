import { render, screen, waitFor } from '@testing-library/react';
import { PostDetail } from './PostDetail';
import { postService } from '../../services/postService';

jest.mock('../../services/postService');
jest.mock('../CommentList/CommentList', () => ({
    __esModule: true,
    default: () => <div>CommentList Mock</div>
}));
jest.mock('../CommentForm/CommentForm', () => ({
    CommentForm: () => <div>CommentForm Mock</div>
}));

const mockedPostService = postService as jest.Mocked<typeof postService>;

describe('PostDetail Component', () => {
    const mockOnBack = jest.fn();

    beforeEach(() => {
        jest.clearAllMocks();
    });

    test('shows loading while the post is being fetched', () => {
        mockedPostService.getPostById.mockImplementation(
            () => new Promise(() => {}) // Never resolves
        );

        render(<PostDetail postId={1} userId={1} onBack={mockOnBack} />);

        expect(screen.getByText('Loading post...')).toBeInTheDocument();
    });

    test('renders the post correctly', async () => {
        const mockPost = {
            id: 1,
            title: 'Test Post',
            content: 'Test Content',
            user_id: 1,
            username: 'testuser',
            created_at: '2024-01-01T00:00:00Z'
        };

        mockedPostService.getPostById.mockResolvedValueOnce(mockPost);

        render(<PostDetail postId={1} userId={1} onBack={mockOnBack} />);

        await waitFor(() => {
            expect(screen.getByText('Test Post')).toBeInTheDocument();
            expect(screen.getByText('Test Content')).toBeInTheDocument();
            expect(screen.getByText(/testuser/i)).toBeInTheDocument();
        });
    });

    test('shows error when loading fails', async () => {
        mockedPostService.getPostById.mockRejectedValueOnce(new Error('Error'));

        render(<PostDetail postId={1} userId={1} onBack={mockOnBack} />);

        await waitFor(() => {
            expect(screen.getByText('Failed to load post')).toBeInTheDocument();
        });
    });

    test('renders CommentForm and CommentList', async () => {
        const mockPost = {
            id: 1,
            title: 'Test Post',
            content: 'Test Content',
            user_id: 1,
            username: 'testuser',
            created_at: '2024-01-01T00:00:00Z'
        };

        mockedPostService.getPostById.mockResolvedValueOnce(mockPost);

        render(<PostDetail postId={1} userId={1} onBack={mockOnBack} />);

        await waitFor(() => {
            expect(screen.getByText('CommentForm Mock')).toBeInTheDocument();
            expect(screen.getByText('CommentList Mock')).toBeInTheDocument();
        });
    });
});