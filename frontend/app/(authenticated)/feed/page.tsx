'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { getFeed, getAvatarUrl } from '@/lib/api';
import { clearAuth } from '@/lib/auth';
import Card from '@/components/ui/Card';

interface FeedPost {
  id: string;
  created_at: string;
  visibility: string;
  author_id: string;
  author_nickname: string;
  author_first_name: string;
  author_last_name: string;
  author_avatar_url: string;
  author_avatar_thumb_url: string;
  title: string;
  content: string;
  images: Array<{
    image_id: string;
    url: string;
    thumbnail_url: string;
    display_order: number;
  }>;
  like_count: number;
  dislike_count: number;
  comment_count: number;
  viewer_reaction: number | null;
}

function formatPostDate(isoDate: string): string {
  if (!isoDate) return '';
  try {
    const d = new Date(isoDate);
    return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
  } catch {
    return isoDate;
  }
}

export default function FeedPage() {
  const router = useRouter();
  const [posts, setPosts] = useState<FeedPost[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    getFeed()
      .then((data: any) => {
        if (cancelled) return;
        // The API returns { posts: [...] }
        const feedPosts = data?.posts || data || [];
        setPosts(feedPosts);
      })
      .catch((err: any) => {
        if (cancelled) return;
        const message = err?.message || 'Failed to load feed';
        console.error('Failed to load feed:', err);
        
        // Redirect to login if authentication is required
        if (message === 'Authentication required' || message.toLowerCase().includes('authentication')) {
          clearAuth();
          router.replace('/login');
          return;
        }
        
        setError(message);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [router]);

  if (loading) {
    return (
      <div className="min-h-screen bg-parea-white p-8">
        <div className="max-w-4xl mx-auto">
          <h1 className="text-4xl font-bold text-parea-black mb-4">Feed</h1>
          <p className="text-regular text-parea-black">Loading feed...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-parea-white p-8">
        <div className="max-w-4xl mx-auto">
          <h1 className="text-4xl font-bold text-parea-black mb-4">Feed</h1>
          <p className="text-regular text-parea-black">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-parea-white p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold text-parea-black mb-4">Feed</h1>
        {posts.length === 0 ? (
          <p className="text-regular text-parea-black">
            No posts yet. Create a post or follow others to see posts here.
          </p>
        ) : (
          <div className="flex flex-col items-start gap-0">
            {posts.map((post, index) => {
              const avatarUrl = post.author_avatar_thumb_url || post.author_avatar_url;
              const imageUrl = post.images?.[0]?.thumbnail_url || post.images?.[0]?.url;

              return (
                <Card
                  key={post.id}
                  imageType="post"
                  imageSrc={imageUrl}
                  avatarSrc={avatarUrl || '/user-avatar-default.png'}
                  avatarAlt={post.author_nickname || 'User'}
                  userName={(post.author_nickname || 'User').toUpperCase()}
                  userDate={formatPostDate(post.created_at)}
                  title={post.title}
                  content={post.content}
                  href={`/post/${post.id}`}
                  imagePriority={index === 0}
                  likeCount={post.like_count}
                  commentCount={post.comment_count}
                />
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
