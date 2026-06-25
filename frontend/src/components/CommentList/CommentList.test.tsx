import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import  CommentList from './CommentList';
import axios from 'axios';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('CommentList Component', () => {
  const mockComments = [
    {
      id: 1,
      post_id: 1,
      user_id: 1,
      username: 'testuser',
      content: 'Mi comentario',
      created_at: '2025-01-01T10:00:00Z'
    },
    {
      id: 2,
      post_id: 1,
      user_id: 2,
      username: 'otheruser',
      content: 'Otro comentario',
      created_at: '2025-01-02T10:00:00Z'
    }
  ];

  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders the comment list correctly', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockComments });

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi comentario')).toBeInTheDocument();
      expect(screen.getByText('Otro comentario')).toBeInTheDocument();
    });

    expect(screen.getByText('Comments (2)')).toBeInTheDocument();
  });

  test('shows "No hay comentarios" when empty', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: [] });

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText(/no comments yet/i)).toBeInTheDocument();
    });
  });

  test('shows delete button only for own comments', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockComments });

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi comentario')).toBeInTheDocument();
    });

    // There should be exactly 1 delete button (for user 1's comment)
    const deleteButtons = screen.queryAllByText(/^delete$/i);
    expect(deleteButtons).toHaveLength(1);
  });

  test('deletes a comment when delete button is clicked', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockComments });
    mockedAxios.delete.mockResolvedValueOnce({ data: {} });

    const mockOnCommentDeleted = jest.fn();

    render(
      <CommentList
        postId={1}
        currentUserId={1}
        onCommentDeleted={mockOnCommentDeleted}
      />
    );

    await waitFor(() => {
      expect(screen.getByText('Mi comentario')).toBeInTheDocument();
    });

    // Click delete
    const deleteButton = screen.getByText(/^delete$/i);
    fireEvent.click(deleteButton);

    // Verify delete was called
    await waitFor(() => {
      expect(mockedAxios.delete).toHaveBeenCalledWith(
        'http://localhost:8080/api/posts/1/comments/1',
        {
          headers: {
            'X-User-ID': '1'
          }
        }
      );
    });

    // Verify callback was called
    expect(mockOnCommentDeleted).toHaveBeenCalledWith(1);
  });

  test('shows error when loading comments fails', async () => {
    mockedAxios.get.mockRejectedValueOnce({
      response: {
        data: {
          error: 'Error del servidor'
        }
      }
    });

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Failed to load comments')).toBeInTheDocument();
    });
  });

  test('should show alert when comment deletion fails', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockComments });
    mockedAxios.delete.mockRejectedValueOnce(new Error('Network error'));
    window.alert = jest.fn();

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi comentario')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/^delete$/i));

    await waitFor(() => {
      expect(window.alert).toHaveBeenCalledWith('Failed to delete comment');
    });
  });

  test('should show success message after comment is deleted', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockComments });
    mockedAxios.delete.mockResolvedValueOnce({ data: {} });

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi comentario')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/^delete$/i));

    await waitFor(() => {
      expect(screen.getByText('Comment deleted successfully')).toBeInTheDocument();
    });
  });
});