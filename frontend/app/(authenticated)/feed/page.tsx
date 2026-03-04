'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { getFeed, getMyGroups, getGroupPosts, getPostById, getAvatarUrl } from '@/lib/api';
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

interface GroupPost {
  id: string;
  user_id: string;
  title: string;
  content: string;
  created_at: string;
  nickname: string;
  thumbnail_url?: string;
  image_url?: string;
  group_id: string;
  group_title: string;
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
  const [groupPosts, setGroupPosts] = useState<GroupPost[]>([]);
  const [groupPostAvatarById, setGroupPostAvatarById] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    if (activeTab === 'Posts') {
      setLoading(true);
      getFeed()
        .then((data: any) => {
          if (cancelled) return;
          const feedPosts = data?.posts ?? data ?? [];
          setPosts(Array.isArray(feedPosts) ? feedPosts : []);
        })
        .catch((err: any) => {
          if (cancelled) return;
          const message = err?.message || 'Failed to load feed';
          console.error('Failed to load feed:', err);
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
      getMyGroups()
        .then(async (groups: any[]) => {
          if (cancelled) return;
          if (!Array.isArray(groups) || groups.length === 0) {
            setGroupPosts([]);
            return;
          }
          const postsByGroup = await Promise.all(
            groups.map(async (group: any) => {
              try {
                const posts = await getGroupPosts(group.id);
                return (Array.isArray(posts) ? posts : []).map((post: any) => ({
                  ...post,
                  group_id: group.id,
                  group_title: group.title || 'Group',
                }));
              } catch {
                return [];
              }
            }),
          );
          if (cancelled) return;
          const all = postsByGroup
            .flat()
            .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
          setGroupPosts(all);
        })
        .catch((err: any) => {
          if (cancelled) return;
          console.error('Failed to load group posts:', err);
          setError(err?.message || 'Failed to load group posts');
        })
        .finally(() => {
          if (!cancelled) setLoading(false);
        });
    }

    return () => {
      cancelled = true;
    };
  }, [router, activeTab]);

  useEffect(() => {
    if (activeTab !== 'Groups' || groupPosts.length === 0) {
      setGroupPostAvatarById({});
      return;
    }

    let cancelled = false;

    (async () => {
      try {
        const entries = await Promise.all(
          groupPosts.map(async (post) => {
            try {
              const detail = await getPostById(post.id);
              const rawAvatar =
                detail?.author_avatar_thumb_url ||
                detail?.author_avatar_url ||
                '';
              const avatar = rawAvatar
                ? (/^https?:\/\//i.test(rawAvatar) ? rawAvatar : getAvatarUrl(rawAvatar))
                : '/user-avatar-default.png';
              return [post.id, avatar] as const;
            } catch {
              return [post.id, '/user-avatar-default.png'] as const;
            }
          }),
        );

        if (cancelled) return;
        const map: Record<string, string> = {};
        for (const [postId, avatar] of entries) {
          map[postId] = avatar;
        }
        setGroupPostAvatarById(map);
      } catch {
        if (!cancelled) setGroupPostAvatarById({});
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [activeTab, groupPosts]);

  return (
    <div className="min-h-screen bg-parea-white p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold text-parea-black mb-4">Feed</h1>
        <Tabs tabs={['Posts', 'Groups']} defaultTab={activeTab} onTabChange={setActiveTab} />

        <div className="mt-6">
          {loading && <p className="text-regular text-parea-black">Loading...</p>}
          {!loading && error && <p className="text-regular text-parea-black">{error}</p>}

          {!loading && !error && activeTab === 'Posts' && (
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
                        userHref={`/profile/${post.author_id}`}
                        title={post.title}
                        content={post.content}
                        href={`/post/${post.id}`}
                        imagePriority={index === 0}
                        commentCount={post.comment_count}
                      />
                    );
                  })}
                </div>
              )}
            </>
          )}

          {!loading && !error && activeTab === 'Groups' && (
            <>
              {groupPosts.length === 0 ? (
                <p className="text-regular text-parea-black">
                  No group posts yet. Join a group or wait for members to post.
                </p>
              ) : (
                <div className="flex flex-col items-start gap-0">
                  {groupPosts.map((post, index) => (
                    <Card
                      key={`${post.group_id}-${post.id}`}
                      imageType="post"
                      avatarSrc={groupPostAvatarById[post.id] || '/user-avatar-default.png'}
                      avatarAlt={post.nickname || 'User'}
                      userName={(post.nickname || 'User').toUpperCase()}
                      userDate={`${formatPostDate(post.created_at)} · ${post.group_title}`}
                      title={post.title}
                      content={post.content}
                      href={`/post/${post.id}`}
                      imagePriority={index === 0}
                    />
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
