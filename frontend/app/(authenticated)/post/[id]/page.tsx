'use client';

import { useParams } from 'next/navigation';
import { useState, useEffect } from 'react';
import PostDetailFrame from '@/components/ui/PostDetailFrame';
import type { CommentItem } from '@/components/ui/CommentHolder';
import {
  getPostById,
  createCommentWithImage,
  getCommentsByPostId,
  getProfile,
  getAvatarUrl,
  getPostImageUrl,
} from '@/lib/api';

function formatPostDate(isoDate: string): string {
  if (!isoDate) return '';
  try {
    const d = new Date(isoDate);
    return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
  } catch {
    return isoDate;
  }
}

function mapCommentToItem(c: any): CommentItem {
  const rawCommentImage = c.image_thumbnail_url || c.image_url;
  return {
    id: c.id ?? c.comment_id ?? '',
    userId: c.author_id,
    avatarSrc: c.author_avatar_thumb_url || c.author_avatar_url || '/user-avatar-default.png',
    avatarAlt: c.author_nickname ?? 'User',
    userName: (c.author_nickname ?? 'User').toUpperCase(),
    userDate: formatPostDate(c.created_at),
    text: c.content ?? '',
    likeCount: c.like_count ?? 0,
    imageSrc: getPostImageUrl(rawCommentImage) || rawCommentImage || undefined,
  };
}

export default function PostPage() {
  const params = useParams();
  const postId = typeof params?.id === 'string' ? params.id : '';
  const [post, setPost] = useState<any | null>(null);
  const [comments, setComments] = useState<CommentItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [commentText, setCommentText] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [commentImage, setCommentImage] = useState<File | null>(null);
  const [commentImagePreview, setCommentImagePreview] = useState<string | null>(null);
  const [currentUserAvatar, setCurrentUserAvatar] = useState<string | null>(null);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);

  useEffect(() => {
    if (!postId) {
      setError('Invalid post');
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    setError(null);
    getPostById(postId)
      .then((data) => {
        if (cancelled) return;
        setPost(data);
        // Load comments for the post
        return getCommentsByPostId(postId).then((commentsList) => {
          if (!cancelled) {
            const mappedComments = Array.isArray(commentsList) ? commentsList.map(mapCommentToItem) : [];
            setComments(mappedComments);
          }
        }).catch((commentsErr) => {
          if (!cancelled) {
            console.warn('Failed to load comments:', commentsErr);
            // Don't fail the whole page if comments fail to load
            setComments([]);
          }
        });
      })
      .catch((err: any) => {
        if (cancelled) return;
        const msg = err?.message ?? 'Failed to load post';
        setError(msg.includes('404') || msg.includes('Not Found') ? 'Post not found' : msg);
        setPost(null);
        setComments([]);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, [postId]);

  useEffect(() => {
    return () => {
      if (commentImagePreview) {
        URL.revokeObjectURL(commentImagePreview);
      }
    };
  }, [commentImagePreview]);

  useEffect(() => {
    let cancelled = false;

    getProfile()
      .then((profile: any) => {
        if (cancelled || !profile) return;
        if (profile.user?.id) {
          setCurrentUserId(profile.user.id);
        }
        const rawPath =
          profile.avatar?.thumbnail_path ||
          profile.avatar?.file_path ||
          undefined;
        if (!rawPath) {
          setCurrentUserAvatar(null);
          return;
        }
        setCurrentUserAvatar(getAvatarUrl(rawPath));
      })
      .catch(() => {
        if (!cancelled) {
          setCurrentUserAvatar(null);
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const handleCommentImageSelect = (file: File) => {
    if (commentImagePreview) {
      URL.revokeObjectURL(commentImagePreview);
    }
    setCommentImage(file);
    setCommentImagePreview(URL.createObjectURL(file));
  };

  const clearCommentImage = () => {
    if (commentImagePreview) {
      URL.revokeObjectURL(commentImagePreview);
    }
    setCommentImage(null);
    setCommentImagePreview(null);
  };

  const handleCommentSubmit = async () => {
    if (!commentText.trim() || submitting || !postId) return;

    setSubmitting(true);
    try {
      const result = await createCommentWithImage(postId, commentText.trim(), commentImage || undefined);
      
      // Add the new comment to the list
      const newComment: CommentItem = {
        id: result.comment?.id || result.comment?.comment_id || result.id || Date.now().toString(),
        userId: currentUserId || undefined,
        avatarSrc: currentUserAvatar || '/user-avatar-default.png',
        avatarAlt: 'You',
        userName: 'YOU',
        userDate: formatPostDate(new Date().toISOString()),
        text: commentText.trim(),
        likeCount: 0,
        imageSrc: commentImagePreview || undefined,
      };
      
      setComments((prev) => [newComment, ...prev]);
      setCommentText('');
      clearCommentImage();
      
      // Update comment count in post
      if (post) {
        setPost({ ...post, comment_count: (post.comment_count || 0) + 1 });
      }
    } catch (err: any) {
      console.error('Failed to create comment:', err);
      // Optionally show error to user
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-parea-white px-16 py-12">
        <div className="max-w-4xl mx-auto">
          <p className="text-regular text-parea-black">Loading post...</p>
        </div>
      </div>
    );
  }

  if (error || !post) {
    return (
      <div className="min-h-screen bg-parea-white px-16 py-12">
        <div className="max-w-4xl mx-auto">
          <p className="text-regular text-parea-black">{error ?? 'Post not found'}</p>
        </div>
      </div>
    );
  }

  // Backend returns FeedPost structure with different field names
  const avatarUrl = post.author_avatar_thumb_url || post.author_avatar_url || '/user-avatar-default.png';
  const postImage = post.images?.[0];
  const imageSrc = postImage?.url || postImage?.thumbnail_url;

  return (
    <div className="min-h-screen bg-parea-white px-16 py-12">
      <div className="w-full">
        <PostDetailFrame
          title={post.title ?? ''}
          avatarSrc={avatarUrl}
          avatarAlt={post.author_nickname ?? 'Author'}
          userId={post.author_id}
          userName={(post.author_nickname ?? 'User').toUpperCase()}
          userDate={formatPostDate(post.created_at)}
          likeCount={post.like_count ?? 0}
          commentCount={post.comment_count ?? 0}
          imageSrc={imageSrc}
          postText={post.content ?? ''}
          comments={comments}
          commentValue={commentText}
          onCommentChange={setCommentText}
          onCommentSubmit={handleCommentSubmit}
          onCommentImageSelect={handleCommentImageSelect}
          onCommentImageRemove={clearCommentImage}
          commentImagePreview={commentImagePreview}
          commentImageName={commentImage?.name || null}
        />
      </div>
    </div>
  );
}
