import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { PostList } from './PostList';
import axios from 'axios';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('PostList Component', () => {
  const mockPosts = [
    {
      id: 1,
      title: 'Mi primer post',
      content: 'Este es el contenido del primer post',
      user_id: 1,
      username: 'testuser',
      created_at: '2025-01-01T10:00:00Z'
    },
    {
      id: 2,
      title: 'Post de otro usuario',
      content: 'Este es de otro usuario',
      user_id: 2,
      username: 'otheruser',
      created_at: '2025-01-02T10:00:00Z'
    }
  ];

  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders the post list correctly', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockPosts });

    render(<PostList currentUserId={1} />);

    // Wait for posts to load
    await waitFor(() => {
      expect(screen.getByText('Mi primer post')).toBeInTheDocument();
      expect(screen.getByText('Post de otro usuario')).toBeInTheDocument();
    });

    // Verify content is displayed
    expect(screen.getByText('Este es el contenido del primer post')).toBeInTheDocument();
    expect(screen.getByText(/by @testuser/)).toBeInTheDocument();
    expect(screen.getByText(/by @otheruser/)).toBeInTheDocument();
  });

  test('shows "No hay posts" when the list is empty', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: [] });

    render(<PostList currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText(/no posts yet/i)).toBeInTheDocument();
    });
  });

  test('shows delete button only for own posts', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockPosts });

    render(<PostList currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi primer post')).toBeInTheDocument();
    });

    // There should be one delete button (for user 1's post)
    const deleteButtons = screen.getAllByText('Delete');
    expect(deleteButtons).toHaveLength(1);

    // User 2's post should not have a delete button
  });

  test('deletes a post when delete button is clicked', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockPosts });
    mockedAxios.delete.mockResolvedValueOnce({ data: {} });
    mockedAxios.get.mockResolvedValueOnce({ data: [] }); // Second call after deletion

    window.confirm = jest.fn(() => true); // Mock confirm dialog

    render(<PostList currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi primer post')).toBeInTheDocument();
    });

    // Click delete
    const deleteButton = screen.getByText('Delete');
    fireEvent.click(deleteButton);

    // Verify delete was called with the correct parameters
    await waitFor(() => {
      expect(mockedAxios.delete).toHaveBeenCalledWith(
        'http://localhost:8080/api/posts/1',
        {
          headers: {
            'X-User-ID': '1'
          }
        }
      );
    });
  });

  test('shows error when loading posts fails', async () => {
    mockedAxios.get.mockRejectedValueOnce({
      response: {
        data: {
          error: 'Error en el servidor'
        }
      }
    });

    render(<PostList currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Failed to load posts')).toBeInTheDocument();
    });
  });

  test('should not call delete api when user cancels confirmation', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockPosts });
    window.confirm = jest.fn(() => false);

    render(<PostList currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi primer post')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Delete'));

    expect(mockedAxios.delete).not.toHaveBeenCalled();
  });

  test('should show alert when delete request fails', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockPosts });
    mockedAxios.delete.mockRejectedValueOnce(new Error('Network error'));
    window.confirm = jest.fn(() => true);
    window.alert = jest.fn();

    render(<PostList currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi primer post')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Delete'));

    await waitFor(() => {
      expect(window.alert).toHaveBeenCalledWith('Failed to delete post');
    });
  });
});