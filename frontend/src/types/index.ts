// Types matching the backend models

export interface User {
    id: number;
    email: string;
    username: string;
    created_at: string;
  }

  export interface Post {
    id: number;
    title: string;
    content: string;
    user_id: number;
    username: string;
    created_at: string;
  }

  export interface Comment {
    id: number;
    post_id: number;
    user_id: number;
    username: string;
    content: string;
    created_at: string;
  }

  export interface LoginRequest {
    email: string;
    password: string;
  }

  export interface RegisterRequest {
    email: string;
    password: string;
    username: string;
  }

  export interface CreatePostRequest {
    title: string;
    content: string;
  }

  export interface CreateCommentRequest {
    content: string;
  }
