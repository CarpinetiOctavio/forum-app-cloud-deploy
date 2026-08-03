import React, { useEffect, useState } from 'react';
import { postService } from '../../services/postService';
import { Post } from '../../types';
import  CommentList  from '../CommentList/CommentList';
import { CommentForm } from '../CommentForm/CommentForm';
import './PostDetail.css';


interface PostDetailProps {
  postId: number;
  userId: number;
  onBack: () => void;
}

export const PostDetail: React.FC<PostDetailProps> = ({ postId, userId, onBack }) => {
  const [post, setPost] = useState<Post | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [refreshComments, setRefreshComments] = useState(0);
  const [isEditing, setIsEditing] = useState(false);
  const [editTitle, setEditTitle] = useState('');
  const [editContent, setEditContent] = useState('');
  const [editError, setEditError] = useState('');
  const [editLoading, setEditLoading] = useState(false);

  useEffect(() => {
    const loadPost = async () => {
      try {
        setLoading(true);
        const data = await postService.getPostById(postId);
        setPost(data);
        setError('');
      } catch (err: any) {
        setError('Failed to load post');
      } finally {
        setLoading(false);
      }
    };

    loadPost();
  }, [postId]);

  const handleCommentCreated = () => {
    setRefreshComments(prev => prev + 1);
  };

  const handleEditClick = () => {
    if (!post) return;
    setEditTitle(post.title);
    setEditContent(post.content);
    setEditError('');
    setIsEditing(true);
  };

  const handleEditCancel = () => {
    setIsEditing(false);
    setEditError('');
  };

  const handleEditSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setEditError('');
    setEditLoading(true);

    try {
      const updated = await postService.editPost(
        postId,
        { title: editTitle, content: editContent },
        userId
      );
      setPost(updated);
      setIsEditing(false);
    } catch (err: any) {
      setEditError(err.response?.data?.error || 'Failed to edit post');
    } finally {
      setEditLoading(false);
    }
  };

  if (loading) {
    return <div className="post-detail-loading">Loading post...</div>;
  }

  if (error || !post) {
    return (
      <div className="post-detail-error">
        <p>{error || 'Post not found'}</p>
        <button onClick={onBack}>Back</button>
      </div>
    );
  }

  return (
    <div className="post-detail-container">
      <button className="back-btn" onClick={onBack}>
        ← Back
      </button>

      <div className="post-detail-card">
        {isEditing ? (
          <form className="post-detail-edit-form" onSubmit={handleEditSubmit}>
            <input
              type="text"
              value={editTitle}
              onChange={(e) => setEditTitle(e.target.value)}
              placeholder="Title"
              required
              disabled={editLoading}
            />
            <textarea
              value={editContent}
              onChange={(e) => setEditContent(e.target.value)}
              placeholder="Content"
              rows={5}
              required
              disabled={editLoading}
            />

            {editError && <div className="error-message">{editError}</div>}

            <div className="post-detail-edit-actions">
              <button type="submit" disabled={editLoading}>
                {editLoading ? 'Saving...' : 'Save'}
              </button>
              <button type="button" onClick={handleEditCancel} disabled={editLoading}>
                Cancel
              </button>
            </div>
          </form>
        ) : (
          <>
            <div className="post-detail-title-row">
              <h1>{post.title}</h1>
              {post.user_id === userId && (
                <button className="edit-btn" onClick={handleEditClick}>
                  Edit
                </button>
              )}
            </div>
            <div className="post-detail-meta">
              <span className="post-detail-author">By @{post.username}</span>
              <span className="post-detail-date">
                {new Date(post.created_at).toLocaleDateString()}
              </span>
            </div>
            <p className="post-detail-content">{post.content}</p>
          </>
        )}
      </div>

      <CommentForm 
        postId={postId} 
        userId={userId} 
        onCommentCreated={handleCommentCreated} 
      />

      <CommentList
        postId={postId}
        currentUserId={userId}
        refreshTrigger={refreshComments}
      />
    </div>
  );
};