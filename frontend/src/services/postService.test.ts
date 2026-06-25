import axios from 'axios';
import { postService, deleteComment } from './postService';

jest.mock('axios');
const mockedAxios = axios as jest.Mocked<typeof axios>;

describe('postService', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    describe('getAllPosts', () => {
        test('obtiene todos los posts correctamente', async () => {
            const mockPosts = [
                { id: 1, title: 'Post 1', content: 'Content 1', user_id: 1, username: 'user1', created_at: '2024-01-01' },
                { id: 2, title: 'Post 2', content: 'Content 2', user_id: 2, username: 'user2', created_at: '2024-01-02' }
            ];
            mockedAxios.get.mockResolvedValueOnce({ data: mockPosts });

            const result = await postService.getAllPosts();

            expect(mockedAxios.get).toHaveBeenCalledWith('http://localhost:8080/api/posts');
            expect(result).toEqual(mockPosts);
        });
    });

    describe('getPostById', () => {
        test('obtiene un post por ID correctamente', async () => {
            const mockPost = { id: 1, title: 'Post 1', content: 'Content 1', user_id: 1, username: 'user1', created_at: '2024-01-01' };
            mockedAxios.get.mockResolvedValueOnce({ data: mockPost });

            const result = await postService.getPostById(1);

            expect(mockedAxios.get).toHaveBeenCalledWith('http://localhost:8080/api/posts/1');
            expect(result).toEqual(mockPost);
        });
    });

    describe('createPost', () => {
        test('crea un post correctamente', async () => {
            const newPost = { title: 'New Post', content: 'New Content' };
            const mockResponse = { id: 1, ...newPost, user_id: 1, username: 'user1', created_at: '2024-01-01' };
            mockedAxios.post.mockResolvedValueOnce({ data: mockResponse });

            const result = await postService.createPost(newPost, 1);

            expect(mockedAxios.post).toHaveBeenCalledWith(
                'http://localhost:8080/api/posts',
                newPost,
                { headers: { 'X-User-ID': '1' } }
            );
            expect(result).toEqual(mockResponse);
        });
    });

    describe('deletePost', () => {
        test('elimina un post correctamente', async () => {
            mockedAxios.delete.mockResolvedValueOnce({ data: {} });

            await postService.deletePost(1, 1);

            expect(mockedAxios.delete).toHaveBeenCalledWith(
                'http://localhost:8080/api/posts/1',
                { headers: { 'X-User-ID': '1' } }
            );
        });
    });

    describe('createComment', () => {
        test('crea un comentario correctamente', async () => {
            const comment = { content: 'Great post!' };
            const mockResponse = {
                id: 1,
                post_id: 1,
                user_id: 1,
                username: 'user1',
                content: 'Great post!',
                created_at: '2024-01-01'
            };
            mockedAxios.post.mockResolvedValueOnce({ data: mockResponse });

            const result = await postService.createComment(1, comment, 1);

            expect(mockedAxios.post).toHaveBeenCalledWith(
                'http://localhost:8080/api/posts/1/comments',
                comment,
                { headers: { 'X-User-ID': '1' } }
            );
            expect(result).toEqual(mockResponse);
        });
    });

    describe('getComments', () => {
        test('obtiene comentarios de un post', async () => {
            const mockComments = [
                { id: 1, post_id: 1, user_id: 1, username: 'user1', content: 'Comment 1', created_at: '2024-01-01' },
                { id: 2, post_id: 1, user_id: 2, username: 'user2', content: 'Comment 2', created_at: '2024-01-02' }
            ];
            mockedAxios.get.mockResolvedValueOnce({ data: mockComments });

            const result = await postService.getComments(1);

            expect(mockedAxios.get).toHaveBeenCalledWith('http://localhost:8080/api/posts/1/comments');
            expect(result).toEqual(mockComments);
        });
    });
});

describe('deleteComment', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    test('elimina un comentario correctamente', async () => {
        mockedAxios.delete.mockResolvedValueOnce({ data: {} });

        await deleteComment(1, 5, 1);

        expect(mockedAxios.delete).toHaveBeenCalledWith(
            'http://localhost:8080/api/posts/1/comments/5',
            { headers: { 'X-User-ID': '1' } }
        );
    });
});