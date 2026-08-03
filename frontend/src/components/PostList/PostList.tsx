import React, { useEffect, useState } from 'react';
import { postService } from '../../services/postService';
import { Post } from '../../types';
import './PostList.css';

interface PostListProps {
    currentUserId: number;
    onRefresh?: boolean;
    onViewPost?: (postId: number) => void;
}

export const PostList: React.FC<PostListProps> = ({ currentUserId, onRefresh, onViewPost }) => {
    const [posts, setPosts] = useState<Post[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    const loadPosts = async () => {
        try {
            setLoading(true);
            const data = await postService.getAllPosts();
            setPosts(data);
            setError('');
        } catch (err: any) {
            setError('Failed to load posts');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadPosts();
    }, [onRefresh]);

    const handleDelete = async (postId: number) => {
        if (!window.confirm('Are you sure you want to delete this post?')) {
            return;
        }

        try {
            await postService.deletePost(postId, currentUserId);
            loadPosts();
        } catch (err: any) {
            alert(err.response?.data?.error || 'Failed to delete post');
        }
    };

    if (loading) {
        return <div className="loading">Loading posts...</div>;
    }

    if (error) {
        return <div className="error">{error}</div>;
    }

    if (posts.length === 0) {
        return <div className="no-posts">No posts yet. Be the first to create one!</div>;
    }

    return (
        <div className="post-list">
            <h2>Posts</h2>
            {posts.map((post) => (
                <div
                    key={post.id}
                    className="post-card"
                    onClick={() => onViewPost && onViewPost(post.id)}
                    style={{ cursor: onViewPost ? 'pointer' : 'default' }}
                >
                    <div className="post-header">
                        <h3>{post.title}</h3>
                        <span className="post-author">by @{post.username}</span>
                    </div>
                    <p className="post-content">{post.content}</p>
                    <div className="post-footer">
            <span className="post-date">
              {new Date(post.created_at).toLocaleDateString()}
            </span>
                        {post.user_id === currentUserId && (
                            <button
                                className="delete-btn"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    handleDelete(post.id);
                                }}
                            >
                                Delete
                            </button>
                        )}
                    </div>
                </div>
            ))}
        </div>
    );
};
