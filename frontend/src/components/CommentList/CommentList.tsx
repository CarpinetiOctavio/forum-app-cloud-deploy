import React, { useEffect, useState, useCallback } from 'react';
import { postService, deleteComment } from '../../services/postService';
import { Comment } from '../../types';
import './CommentList.css';

interface CommentListProps {
    postId: number;
    currentUserId: number;
    refreshTrigger?: number;
    onCommentDeleted?: (commentId: number) => void;
}

const CommentList: React.FC<CommentListProps> = ({
    postId,
    currentUserId,
    refreshTrigger,
    onCommentDeleted
}) => {
    const [comments, setComments] = useState<Comment[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');
    const [successMessage, setSuccessMessage] = useState('');
    const [editingCommentId, setEditingCommentId] = useState<number | null>(null);
    const [editContent, setEditContent] = useState('');
    const [editError, setEditError] = useState('');
    const [editLoading, setEditLoading] = useState(false);

    // Memoize loadComments to avoid ESLint warning
    const loadComments = useCallback(async () => {
        try {
            setLoading(true);
            const data = await postService.getComments(postId);
            setComments(data);
            setError('');
        } catch (err: any) {
            setError('Failed to load comments');
        } finally {
            setLoading(false);
        }
    }, [postId]);

    useEffect(() => {
        loadComments();
    }, [loadComments, refreshTrigger]);

    const handleDelete = async (commentId: number) => {
        try {
            await deleteComment(postId, commentId, currentUserId);

            // Update comment list using state updater function
            setComments(prev => prev.filter(c => c.id !== commentId));

            if (onCommentDeleted) onCommentDeleted(commentId);

            // Show success message
            setSuccessMessage('Comment deleted successfully');
            setTimeout(() => setSuccessMessage(''), 3000);

        } catch (err: any) {
            console.error("Error deleting comment:", err.response?.data || err.message);
            alert("Failed to delete comment");
        }
    };

    const handleEditClick = (comment: Comment) => {
        setEditingCommentId(comment.id);
        setEditContent(comment.content);
        setEditError('');
    };

    const handleEditCancel = () => {
        setEditingCommentId(null);
        setEditContent('');
        setEditError('');
    };

    const handleEditSubmit = async (e: React.FormEvent, commentId: number) => {
        e.preventDefault();
        setEditError('');
        setEditLoading(true);

        try {
            const updated = await postService.editComment(postId, commentId, { content: editContent }, currentUserId);
            setComments(prev => prev.map(c => (c.id === commentId ? updated : c)));
            setEditingCommentId(null);
        } catch (err: any) {
            setEditError(err.response?.data?.error || 'Failed to edit comment');
        } finally {
            setEditLoading(false);
        }
    };

    if (loading) return <div className="comments-loading">Loading comments...</div>;
    if (error) return <div className="comments-error">{error}</div>;

    if (comments.length === 0) {
        return (
            <div className="comment-list">
                {successMessage && <div className="success-message">{successMessage}</div>}
                <div className="no-comments">No comments yet. Be the first to comment!</div>
            </div>
        );
    }

    return (
        <div className="comment-list">
            <h3>Comments ({comments.length})</h3>

            {successMessage && <div className="success-message">{successMessage}</div>}

            {comments.map(comment => (
                <div key={comment.id} className="comment-card">
                    {editingCommentId === comment.id ? (
                        <form
                            className="comment-edit-form"
                            onSubmit={(e) => handleEditSubmit(e, comment.id)}
                        >
                            <textarea
                                value={editContent}
                                onChange={(e) => setEditContent(e.target.value)}
                                rows={3}
                                required
                                disabled={editLoading}
                            />

                            {editError && <div className="error-message">{editError}</div>}

                            <div className="comment-edit-actions">
                                <button
                                    type="submit"
                                    className="comment-save-btn"
                                    disabled={editLoading}
                                >
                                    {editLoading ? 'Saving...' : 'Save'}
                                </button>
                                <button
                                    type="button"
                                    className="comment-cancel-btn"
                                    onClick={handleEditCancel}
                                    disabled={editLoading}
                                >
                                    Cancel
                                </button>
                            </div>
                        </form>
                    ) : (
                        <>
                            <div className="comment-header">
                                <span className="comment-author">@{comment.username}</span>
                                <span className="comment-date">{new Date(comment.created_at).toLocaleDateString()}</span>
                                {comment.user_id === currentUserId && (
                                    <>
                                        <button
                                            className="comment-edit-btn"
                                            onClick={() => handleEditClick(comment)}
                                        >
                                            Edit
                                        </button>
                                        <button
                                            className="comment-delete-btn"
                                            onClick={() => handleDelete(comment.id)}
                                        >
                                            Delete
                                        </button>
                                    </>
                                )}
                            </div>
                            <p className="comment-content">{comment.content}</p>
                        </>
                    )}
                </div>
            ))}
        </div>
    );
};

export default CommentList;