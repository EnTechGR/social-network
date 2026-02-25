'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { getFeed, getAllGroups, getAvatarUrl } from '@/lib/api';
import { clearAuth } from '@/lib/auth';
import Card from '@/components/ui/Card';
import Tabs from '@/components/ui/Tabs';

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
  const [activeTab, setActiveTab] = useState('Posts');
  const [posts, setPosts] = useState<FeedPost[]>([]);
  const [groups, setGroups] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    if (activeTab === 'Posts') {
      setLoading(true);
      getFeed()
        .then((data: any) => {
          if (cancelled) return;
          // The API returns { posts: [...] }
          const feedPosts = data?.posts ?? data ?? [];
          setPosts(Array.isArray(feedPosts) ? feedPosts : []);
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
    } else if (activeTab === 'Groups') {
      setLoading(true);
      getAllGroups()
        .then((data: any[]) => {
          if (cancelled) return;
          setGroups(data);
        })
        .catch((err: any) => {
          if (cancelled) return;
          const message = err?.message || 'Failed to load groups';
          console.error('Failed to load groups:', err);
          setError(message);
        })
        .finally(() => {
          if (!cancelled) setLoading(false);
        });
    }

    return () => {
      cancelled = true;
    };
  }, [router, activeTab]);

  if (loading) {
    return (
      <div className="min-h-screen bg-parea-white p-8">
        <div className="max-w-4xl mx-auto">
          <h1 className="text-4xl font-bold text-parea-black mb-4">Feed</h1>
          <Tabs tabs={['Posts', 'Groups']} defaultTab={activeTab} onTabChange={setActiveTab} />
          <div className="mt-6">
            <p className="text-regular text-parea-black">Loading...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-parea-white p-8">
        <div className="max-w-4xl mx-auto">
          <h1 className="text-4xl font-bold text-parea-black mb-4">Feed</h1>
          <Tabs tabs={['Posts', 'Groups']} defaultTab={activeTab} onTabChange={setActiveTab} />
          <div className="mt-6">
            <p className="text-regular text-parea-black">{error}</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-parea-white p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold text-parea-black mb-4">Feed</h1>
        <Tabs tabs={['Posts', 'Groups']} defaultTab={activeTab} onTabChange={setActiveTab} />
        
        <div className="mt-6">
          {activeTab === 'Posts' && (
            <>
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
            </>
          )}

          {activeTab === 'Groups' && (
            <>
              {groups.length === 0 ? (
                <p className="text-regular text-parea-black">
                  No groups yet. Create a group to get started.
                </p>
              ) : (
                <div className="flex flex-col gap-4">
                  {groups.map((group) => (
                    <div
                      key={group.id}
                      onClick={() => router.push(`/group/${group.id}`)}
                      className="border border-parea-black p-6 bg-white cursor-pointer hover:shadow-[4px_4px_0_0_#000] transition-shadow"
                    >
                      <h3 className="text-2xl font-bold text-parea-black mb-2">
                        {group.title}
                      </h3>
                      {group.description && (
                        <p className="text-regular text-parea-black/70 mb-3">
                          {group.description}
                        </p>
                      )}
                      <div className="flex gap-4 text-sm text-parea-black/60">
                        <span>{group.member_count || 0} members</span>
                        <span>•</span>
                        <span>Created by {group.owner_nickname || 'Unknown'}</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
