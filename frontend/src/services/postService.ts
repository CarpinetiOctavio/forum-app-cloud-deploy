import axios from 'axios';
import { Post, CreatePostRequest, Comment, CreateCommentRequest } from '../types';

const API_BASE = process.env.REACT_APP_API_URL || 'http://localhost:8080/api';
const API_URL = `${API_BASE}/posts`;

export const postService = {
  // Retrieve all posts
  async getAllPosts(): Promise<Post[]> {
    const response = await axios.get<Post[]>(API_URL);
    return response.data;
  },

  // Create a new post
  async createPost(data: CreatePostRequest, userId: number): Promise<Post> {
    const response = await axios.post<Post>(API_URL, data, {
      headers: {
        'X-User-ID': userId.toString()
      }
    });
    return response.data;
  },

  // Get a post by ID
  async getPostById(id: number): Promise<Post> {
    const response = await axios.get<Post>(`${API_URL}/${id}`);
    return response.data;
  },

  // Delete a post
  async deletePost(id: number, userId: number): Promise<void> {
    await axios.delete(`${API_URL}/${id}`, {
      headers: {
        'X-User-ID': userId.toString()
      }
    });
  },

  // Get comments for a post
  async getComments(postId: number): Promise<Comment[]> {
    const response = await axios.get<Comment[]>(`${API_URL}/${postId}/comments`);
    return response.data;
  },

  // Create a comment
  async createComment(postId: number, data: CreateCommentRequest, userId: number): Promise<Comment> {
    const response = await axios.post<Comment>(
      `${API_URL}/${postId}/comments`,
      data,
      {
        headers: {
          'X-User-ID': userId.toString()
        }
      }
    );
    return response.data;
  }
};

// Delete a comment
export const deleteComment = async (postId: number, commentId: number, userId: number) => {
    return axios.delete(`${API_URL}/${postId}/comments/${commentId}`, {
        headers: { 'X-User-ID': userId.toString() }
    });
};
