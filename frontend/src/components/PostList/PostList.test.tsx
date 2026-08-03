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

  test('shows an empty-state message when there are no posts', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: [] });

    render(<PostList currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText(/no posts yet/i)).toBeInTheDocument();
    });
  });

  test("shows the delete button only for the current user's own posts", async () => {
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

  test('deletes a post when the delete button is clicked', async () => {
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

  test('does not call deletePost when the user cancels the confirmation dialog', async () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockPosts });
    window.confirm = jest.fn(() => false);

    render(<PostList currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Mi primer post')).toBeInTheDocument();
    });

    const deleteButton = screen.getByText('Delete');
    fireEvent.click(deleteButton);

    expect(mockedAxios.delete).not.toHaveBeenCalled();
  });

  test('shows an error message when loading posts fails', async () => {
    // PostList.tsx ignores the backend's error message entirely and always
    // shows its own fixed string, so the mock only needs to simulate a
    // failed request — the rejection's content is never read.
    mockedAxios.get.mockRejectedValueOnce(new Error('Request failed with status code 500'));

    render(<PostList currentUserId={1} />);

    await waitFor(() => {
      expect(screen.getByText('Failed to load posts')).toBeInTheDocument();
    });
  });
});