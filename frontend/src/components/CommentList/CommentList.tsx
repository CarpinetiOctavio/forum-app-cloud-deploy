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

    /* istanbul ignore next */
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

    if (loading) return <div className="comments-loading">Loading comments...</div>;
    if (error) return <div className="comments-error">{error}</div>;
    if (comments.length === 0) return <div className="no-comments">No comments yet. Be the first to comment!</div>;

    return (
        <div className="comment-list">
            <h3>Comments ({comments.length})</h3>

            {successMessage && <div className="success-message">{successMessage}</div>}

            {comments.map(comment => (
                <div key={comment.id} className="comment-card">
                    <div className="comment-header">
                        <span className="comment-author">@{comment.username}</span>
                        <span className="comment-date">{new Date(comment.created_at).toLocaleDateString()}</span>
                        {comment.user_id === currentUserId && (
                            <button
                                className="comment-delete-btn"
                                onClick={() => handleDelete(comment.id)}
                            >
                                Delete
                            </button>
                        )}
                    </div>
                    <p className="comment-content">{comment.content}</p>
                </div>
            ))}
        </div>
    );
};

export default CommentList;
