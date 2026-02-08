'use client';

import { useParams } from 'next/navigation';
import { useState, useEffect } from 'react';
import PostDetailFrame from '@/components/ui/PostDetailFrame';
import type { CommentItem } from '@/components/ui/CommentHolder';
import { getPostById, getCommentsByPostId, getAvatarUrl } from '@/lib/api';

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
  const avatarPath = c.avatar?.file_path || c.avatar?.thumbnail_path;
  return {
    id: c.id ?? c.comment_id ?? '',
    avatarSrc: avatarPath ? getAvatarUrl(avatarPath) : '/user-avatar-default.png',
    avatarAlt: c.nickname ?? 'User',
    userName: (c.nickname ?? 'User').toUpperCase(),
    userDate: formatPostDate(c.created_at),
    text: c.content ?? c.text ?? '',
    likeCount: c.reactions?.length ?? 0,
  };
}

export default function PostPage() {
  const params = useParams();
  const postId = typeof params?.id === 'string' ? params.id : '';
  const [post, setPost] = useState<any | null>(null);
  const [comments, setComments] = useState<CommentItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

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
        return getCommentsByPostId(postId).then((list) => {
          if (!cancelled) setComments(Array.isArray(list) ? list.map(mapCommentToItem) : []);
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

  const avatarPath = post.avatar?.file_path || post.avatar?.thumbnail_path;
  const imageSrc = post.image_url || post.thumbnail_url || '/test-post.png';

  return (
    <div className="min-h-screen bg-parea-white px-16 py-12">
      <div className="w-full">
        <PostDetailFrame
          title={post.title ?? ''}
          avatarSrc={avatarPath ? getAvatarUrl(avatarPath) : '/user-avatar-default.png'}
          avatarAlt={post.nickname ?? 'Author'}
          userName={(post.nickname ?? 'User').toUpperCase()}
          userDate={formatPostDate(post.created_at)}
          likeCount={post.like_count ?? post.reactions?.length ?? 0}
          commentCount={comments.length}
          imageSrc={imageSrc}
          postText={post.content ?? ''}
          comments={comments}
        />
      </div>
    </div>
  );
}
