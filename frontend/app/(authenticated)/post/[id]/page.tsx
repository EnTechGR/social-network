'use client';

import { useParams } from 'next/navigation';
import { useState, useEffect } from 'react';
import PostDetailFrame from '@/components/ui/PostDetailFrame';
import type { CommentItem } from '@/components/ui/CommentHolder';
import { getPostById } from '@/lib/api';

function formatPostDate(isoDate: string): string {
  if (!isoDate) return '';
  try {
    const d = new Date(isoDate);
    return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
  } catch {
    return isoDate;
  }
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
        // Comments endpoint doesn't exist yet, so skip loading comments
        // When backend adds /api/v1/posts/{id}/comments endpoint, uncomment:
        // return getCommentsByPostId(postId).then((list) => {
        //   if (!cancelled) setComments(Array.isArray(list) ? list.map(mapCommentToItem) : []);
        // });
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

  // Backend returns FeedPost structure with different field names
  const avatarUrl = post.author_avatar_thumb_url || post.author_avatar_url || '/user-avatar-default.png';
  const postImage = post.images?.[0];
  const imageSrc = postImage?.thumbnail_url || postImage?.url;

  return (
    <div className="min-h-screen bg-parea-white px-16 py-12">
      <div className="w-full">
        <PostDetailFrame
          title={post.title ?? ''}
          avatarSrc={avatarUrl}
          avatarAlt={post.author_nickname ?? 'Author'}
          userName={(post.author_nickname ?? 'User').toUpperCase()}
          userDate={formatPostDate(post.created_at)}
          likeCount={post.like_count ?? 0}
          commentCount={post.comment_count ?? 0}
          imageSrc={imageSrc}
          postText={post.content ?? ''}
          comments={comments}
        />
      </div>
    </div>
  );
}
