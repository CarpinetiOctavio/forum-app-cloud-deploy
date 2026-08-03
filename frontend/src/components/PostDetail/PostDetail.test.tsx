import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { PostDetail } from './PostDetail';
import axios from 'axios';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('PostDetail Component', () => {
  const mockPost = {
    id: 1,
    title: 'Mi primer post',
    content: 'Este es el contenido del post',
    user_id: 1,
    username: 'testuser',
    created_at: '2025-01-01T10:00:00Z'
  };

  beforeEach(() => {
    jest.clearAllMocks();
  });

  // PostDetail renders CommentList as a child, which fires its own GET for
  // comments on mount -- every test needs a second queued response for that.
  const mockLoadWithNoComments = () => {
    mockedAxios.get.mockResolvedValueOnce({ data: mockPost });
    mockedAxios.get.mockResolvedValueOnce({ data: [] });
  };

  test('renders the post title, content, author and date', async () => {
    mockLoadWithNoComments();

    render(<PostDetail postId={1} userId={1} onBack={jest.fn()} />);

    await waitFor(() => {
      expect(screen.getByText('Mi primer post')).toBeInTheDocument();
    });

    expect(screen.getByText('Este es el contenido del post')).toBeInTheDocument();
    expect(screen.getByText('By @testuser')).toBeInTheDocument();
  });

  test('shows the edit button when the current user is the author', async () => {
    mockLoadWithNoComments();

    render(<PostDetail postId={1} userId={1} onBack={jest.fn()} />);

    await waitFor(() => {
      expect(screen.getByText('Mi primer post')).toBeInTheDocument();
    });

    expect(screen.getByText(/^edit$/i)).toBeInTheDocument();
  });

  test('hides the edit button when the current user is not the author', async () => {
    mockLoadWithNoComments();

    render(<PostDetail postId={1} userId={2} onBack={jest.fn()} />);

    await waitFor(() => {
      expect(screen.getByText('Mi primer post')).toBeInTheDocument();
    });

    expect(screen.queryByText(/^edit$/i)).not.toBeInTheDocument();
  });

  test('edits the post when the edit form is submitted', async () => {
    mockLoadWithNoComments();
    mockedAxios.put.mockResolvedValueOnce({
      data: { ...mockPost, title: 'Titulo editado', content: 'Contenido editado' }
    });

    render(<PostDetail postId={1} userId={1} onBack={jest.fn()} />);

    await waitFor(() => {
      expect(screen.getByText('Mi primer post')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/^edit$/i));

    fireEvent.change(screen.getByDisplayValue('Mi primer post'), {
      target: { value: 'Titulo editado' }
    });
    fireEvent.change(screen.getByDisplayValue('Este es el contenido del post'), {
      target: { value: 'Contenido editado' }
    });
    fireEvent.click(screen.getByText(/^save$/i));

    await waitFor(() => {
      expect(mockedAxios.put).toHaveBeenCalledWith(
        'http://localhost:8080/api/posts/1',
        { title: 'Titulo editado', content: 'Contenido editado' },
        { headers: { 'X-User-ID': '1' } }
      );
    });

    await waitFor(() => {
      expect(screen.getByText('Titulo editado')).toBeInTheDocument();
      expect(screen.getByText('Contenido editado')).toBeInTheDocument();
    });
  });

  test('shows an error and keeps editing open when the edit is rejected', async () => {
    mockLoadWithNoComments();
    mockedAxios.put.mockRejectedValueOnce({
      response: { data: { error: 'you do not have permission to edit this post' } }
    });

    render(<PostDetail postId={1} userId={1} onBack={jest.fn()} />);

    await waitFor(() => {
      expect(screen.getByText('Mi primer post')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/^edit$/i));
    fireEvent.click(screen.getByText(/^save$/i));

    await waitFor(() => {
      expect(screen.getByText('you do not have permission to edit this post')).toBeInTheDocument();
    });

    // Still in edit mode -- the title input is still there
    expect(screen.getByDisplayValue('Mi primer post')).toBeInTheDocument();
  });

  test('cancels editing without calling the API', async () => {
    mockLoadWithNoComments();

    render(<PostDetail postId={1} userId={1} onBack={jest.fn()} />);

    await waitFor(() => {
      expect(screen.getByText('Mi primer post')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText(/^edit$/i));
    fireEvent.click(screen.getByText(/^cancel$/i));

    expect(screen.getByText('Mi primer post')).toBeInTheDocument();
    expect(mockedAxios.put).not.toHaveBeenCalled();
  });

  test('shows an error message when loading the post fails', async () => {
    mockedAxios.get.mockRejectedValueOnce(new Error('Request failed with status code 404'));

    render(<PostDetail postId={999} userId={1} onBack={jest.fn()} />);

    await waitFor(() => {
      expect(screen.getByText('Failed to load post')).toBeInTheDocument();
    });
  });
});
