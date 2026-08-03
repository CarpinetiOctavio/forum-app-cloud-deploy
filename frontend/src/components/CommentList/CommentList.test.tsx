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

  test('shows an empty-state message when there are no comments', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: [] });

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText(/no comments yet/i)).toBeInTheDocument();
    });
  });

  test("shows the delete button only for the current user's own comments", async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockComments });

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi comentario')).toBeInTheDocument();
    });

    // There should be exactly 1 delete button (for user 1's comment)
    const deleteButtons = screen.queryAllByText(/^delete$/i);
    expect(deleteButtons).toHaveLength(1);
  });

  test("shows the edit button only for the current user's own comments", async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockComments });

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi comentario')).toBeInTheDocument();
    });

    // There should be exactly 1 edit button (for user 1's comment)
    const editButtons = screen.queryAllByText(/^edit$/i);
    expect(editButtons).toHaveLength(1);
  });

  test('edits a comment when the edit form is submitted', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockComments });
    mockedAxios.put.mockResolvedValueOnce({
      data: { ...mockComments[0], content: 'Comentario editado' }
    });

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi comentario')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/^edit$/i));

    const textarea = screen.getByDisplayValue('Mi comentario');
    fireEvent.change(textarea, { target: { value: 'Comentario editado' } });
    fireEvent.click(screen.getByText(/^save$/i));

    await waitFor(() => {
      expect(mockedAxios.put).toHaveBeenCalledWith(
        'http://localhost:8080/api/posts/1/comments/1',
        { content: 'Comentario editado' },
        { headers: { 'X-User-ID': '1' } }
      );
    });

    await waitFor(() => {
      expect(screen.getByText('Comentario editado')).toBeInTheDocument();
    });
  });

  test('shows an error and keeps editing open when the edit is rejected', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockComments });
    mockedAxios.put.mockRejectedValueOnce({
      response: { data: { error: 'you do not have permission to edit this comment' } }
    });

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi comentario')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/^edit$/i));
    fireEvent.click(screen.getByText(/^save$/i));

    await waitFor(() => {
      expect(screen.getByText('you do not have permission to edit this comment')).toBeInTheDocument();
    });

    // Still in edit mode -- the textarea is still there
    expect(screen.getByDisplayValue('Mi comentario')).toBeInTheDocument();
  });

  test('cancels editing without calling the API', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockComments });

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi comentario')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/^edit$/i));
    fireEvent.click(screen.getByText(/^cancel$/i));

    expect(screen.getByText('Mi comentario')).toBeInTheDocument();
    expect(mockedAxios.put).not.toHaveBeenCalled();
  });

  test('deletes a comment when the delete button is clicked', async () => {
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

    // Verify the callback was called
    expect(mockOnCommentDeleted).toHaveBeenCalledWith(1);
  });

  test('shows an error message when loading comments fails', async () => {
    mockedAxios.get.mockRejectedValueOnce({
      response: {
        data: {
          error: 'Failed to load comments'
        }
      }
    });

    render(<CommentList postId={1} currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Failed to load comments')).toBeInTheDocument();
    });
  });
});