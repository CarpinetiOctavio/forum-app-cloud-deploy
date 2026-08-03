import axios from 'axios';
import { postService, deleteComment } from './postService';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('postService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('getAllPosts', () => {
    test('calls the API and returns the post list', async () => {
      const mockPosts = [
        { id: 1, title: 'First post', content: 'Content', user_id: 1, username: 'testuser', created_at: '2025-01-01' }
      ];

      mockedAxios.get.mockResolvedValueOnce({ data: mockPosts });

      const result = await postService.getAllPosts();

      expect(mockedAxios.get).toHaveBeenCalledWith('http://localhost:8080/api/posts');
      expect(result).toEqual(mockPosts);
    });

    test('propagates the error when the request fails', async () => {
      const error = new Error('network error');
      mockedAxios.get.mockRejectedValueOnce(error);

      await expect(postService.getAllPosts()).rejects.toEqual(error);
    });
  });

  describe('createPost', () => {
    test('sends title, content and the X-User-ID header', async () => {
      const mockPost = { id: 1, title: 'New post', content: 'Content', user_id: 1, username: 'testuser', created_at: '2025-01-01' };
      mockedAxios.post.mockResolvedValueOnce({ data: mockPost });

      const result = await postService.createPost({ title: 'New post', content: 'Content' }, 1);

      expect(mockedAxios.post).toHaveBeenCalledWith(
        'http://localhost:8080/api/posts',
        { title: 'New post', content: 'Content' },
        { headers: { 'X-User-ID': '1' } }
      );
      expect(result).toEqual(mockPost);
    });

    test('propagates the error when creation is rejected', async () => {
      const error = new Error('title must be at least 3 characters');
      mockedAxios.post.mockRejectedValueOnce(error);

      await expect(
        postService.createPost({ title: 'ab', content: 'Content' }, 1)
      ).rejects.toEqual(error);
    });
  });

  describe('getPostById', () => {
    test('calls the API with the post id and returns the post', async () => {
      const mockPost = { id: 5, title: 'Post 5', content: 'Content', user_id: 1, username: 'testuser', created_at: '2025-01-01' };
      mockedAxios.get.mockResolvedValueOnce({ data: mockPost });

      const result = await postService.getPostById(5);

      expect(mockedAxios.get).toHaveBeenCalledWith('http://localhost:8080/api/posts/5');
      expect(result).toEqual(mockPost);
    });

    test('propagates the error when the post does not exist', async () => {
      const error = new Error('post not found');
      mockedAxios.get.mockRejectedValueOnce(error);

      await expect(postService.getPostById(999)).rejects.toEqual(error);
    });
  });

  describe('editPost', () => {
    test('sends title, content and the X-User-ID header to the correct post', async () => {
      const mockPost = { id: 1, title: 'Updated title', content: 'Updated content', user_id: 1, username: 'testuser', created_at: '2025-01-01' };
      mockedAxios.put.mockResolvedValueOnce({ data: mockPost });

      const result = await postService.editPost(1, { title: 'Updated title', content: 'Updated content' }, 1);

      expect(mockedAxios.put).toHaveBeenCalledWith(
        'http://localhost:8080/api/posts/1',
        { title: 'Updated title', content: 'Updated content' },
        { headers: { 'X-User-ID': '1' } }
      );
      expect(result).toEqual(mockPost);
    });

    test('propagates the error when the requester is not the author', async () => {
      const error = new Error('you do not have permission to edit this post');
      mockedAxios.put.mockRejectedValueOnce(error);

      await expect(
        postService.editPost(1, { title: 'Updated title', content: 'Updated content' }, 2)
      ).rejects.toEqual(error);
    });
  });

  describe('deletePost', () => {
    test('sends the delete request with the X-User-ID header', async () => {
      mockedAxios.delete.mockResolvedValueOnce({ data: undefined });

      await postService.deletePost(1, 1);

      expect(mockedAxios.delete).toHaveBeenCalledWith(
        'http://localhost:8080/api/posts/1',
        { headers: { 'X-User-ID': '1' } }
      );
    });

    test('propagates the error when the requester is not the author', async () => {
      const error = new Error('you do not have permission to delete this post');
      mockedAxios.delete.mockRejectedValueOnce(error);

      await expect(postService.deletePost(1, 2)).rejects.toEqual(error);
    });
  });

  describe('getComments', () => {
    test('calls the API with the post id and returns the comment list', async () => {
      const mockComments = [
        { id: 1, post_id: 1, user_id: 1, username: 'testuser', content: 'A comment', created_at: '2025-01-01' }
      ];
      mockedAxios.get.mockResolvedValueOnce({ data: mockComments });

      const result = await postService.getComments(1);

      expect(mockedAxios.get).toHaveBeenCalledWith('http://localhost:8080/api/posts/1/comments');
      expect(result).toEqual(mockComments);
    });

    test('propagates the error when the parent post does not exist', async () => {
      const error = new Error('post not found');
      mockedAxios.get.mockRejectedValueOnce(error);

      await expect(postService.getComments(999)).rejects.toEqual(error);
    });
  });

  describe('createComment', () => {
    test('sends content and the X-User-ID header on the correct post', async () => {
      const mockComment = { id: 1, post_id: 1, user_id: 1, username: 'testuser', content: 'A comment', created_at: '2025-01-01' };
      mockedAxios.post.mockResolvedValueOnce({ data: mockComment });

      const result = await postService.createComment(1, { content: 'A comment' }, 1);

      expect(mockedAxios.post).toHaveBeenCalledWith(
        'http://localhost:8080/api/posts/1/comments',
        { content: 'A comment' },
        { headers: { 'X-User-ID': '1' } }
      );
      expect(result).toEqual(mockComment);
    });

    test('propagates the error when content is empty', async () => {
      const error = new Error('comment content is required');
      mockedAxios.post.mockRejectedValueOnce(error);

      await expect(
        postService.createComment(1, { content: '' }, 1)
      ).rejects.toEqual(error);
    });
  });

  describe('editComment', () => {
    test('sends content and the X-User-ID header scoped to post and comment', async () => {
      const mockComment = { id: 2, post_id: 1, user_id: 1, username: 'testuser', content: 'Updated comment', created_at: '2025-01-01' };
      mockedAxios.put.mockResolvedValueOnce({ data: mockComment });

      const result = await postService.editComment(1, 2, { content: 'Updated comment' }, 1);

      expect(mockedAxios.put).toHaveBeenCalledWith(
        'http://localhost:8080/api/posts/1/comments/2',
        { content: 'Updated comment' },
        { headers: { 'X-User-ID': '1' } }
      );
      expect(result).toEqual(mockComment);
    });

    test('propagates the error when the requester is not the author', async () => {
      const error = new Error('you do not have permission to edit this comment');
      mockedAxios.put.mockRejectedValueOnce(error);

      await expect(
        postService.editComment(1, 2, { content: 'Updated comment' }, 2)
      ).rejects.toEqual(error);
    });
  });

  describe('deleteComment', () => {
    test('sends the delete request scoped to post, comment and the X-User-ID header', async () => {
      mockedAxios.delete.mockResolvedValueOnce({ data: undefined });

      await deleteComment(1, 2, 1);

      expect(mockedAxios.delete).toHaveBeenCalledWith(
        'http://localhost:8080/api/posts/1/comments/2',
        { headers: { 'X-User-ID': '1' } }
      );
    });

    test('propagates the error when the requester is not the author', async () => {
      const error = new Error('you do not have permission to delete this comment or it does not exist');
      mockedAxios.delete.mockRejectedValueOnce(error);

      await expect(deleteComment(1, 2, 2)).rejects.toEqual(error);
    });
  });
});
