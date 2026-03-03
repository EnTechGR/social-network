/**
 * app/profile/page.tsx
 *
 * Profile page - accessible at /profile
 * Currently using mock data for testing the ProfileCard component.
 * TODO: Replace mock data with real API calls when backend is ready.
 */

'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import ProfileWrap from '@/components/ui/ProfileWrap';
import Tabs from '@/components/ui/Tabs';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import CreateGroupModal from '@/components/ui/CreateGroupModal';
import FollowersModal from '@/components/ui/FollowersModal';
import {
  getProfile,
  updatePrivacy,
  uploadAvatar,
  getAvatarUrl,
  getPostImageUrl,
  getPostsByUserId,
  getPostById,
  getFeed,
  getMyGroups,
  getUserProfile,
  removeFollower,
  unfollowUser,
} from '@/lib/api';
import { clearAuth } from '@/lib/auth';

// When false, API failures show error or redirect to login instead of mock data
const USE_MOCK_FALLBACK = false;

// Mock user data for testing - replace with API call later
const mockUser = {
  avatarUrl: '/test-avatar.png',
  name: 'Olivia Winter',
  username: 'owinter',
  bio: 'Urban explorer, coffee enthusiast, and amateur photographer. Always chasing hidden gems in the city and capturing everyday moments that tell a story. Lover of slow mornings, cozy cafés, and spontaneous adventures.',
  email: 'olivia@mail.com',
  birthDate: '05/10/1994',
  gender: 'prefer_not_to_say',
  isPublic: true,
  followersCount: 356,
  followingCount: 250,
};

type ProfileUser = typeof mockUser;

type RelationUser = {
  user_id: string;
  nickname?: string;
  first_name?: string;
  last_name?: string;
  email?: string;
};


export default function ProfilePage() {
  const router = useRouter();
  const [user, setUser] = useState<ProfileUser | null>(null);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [, setIsLoading] = useState(true);
  const [, setError] = useState<string | null>(null);
  const [isPublic, setIsPublic] = useState(true);
  const [activeTab, setActiveTab] = useState('Posts');
  const [, setIsUploading] = useState(false);
  const [myPosts, setMyPosts] = useState<any[]>([]);
  const [postsLoading, setPostsLoading] = useState(false);
  const [postsError, setPostsError] = useState<string | null>(null);
  const [feedImagesByPostId, setFeedImagesByPostId] = useState<Record<string, string | undefined>>({});
  const [myGroups, setMyGroups] = useState<any[]>([]);
  const [groupsLoading, setGroupsLoading] = useState(false);
  const [groupsError, setGroupsError] = useState<string | null>(null);
  const [createGroupOpen, setCreateGroupOpen] = useState(false);
  const [followers, setFollowers] = useState<RelationUser[]>([]);
  const [following, setFollowing] = useState<RelationUser[]>([]);
  const [, setFollowDataLoading] = useState(false);
  const [followDataError, setFollowDataError] = useState<string | null>(null);
  const [followActionUserId, setFollowActionUserId] = useState<string | null>(null);
  const [showFollowersModal, setShowFollowersModal] = useState(false);
  const [showFollowingModal, setShowFollowingModal] = useState(false);


  const refreshFollowData = useCallback(async (userId: string) => {
    setFollowDataLoading(true);
    setFollowDataError(null);

    try {
      const profileView = await getUserProfile(userId);
      if (profileView && !profileView.privateProfile) {
        setFollowers(Array.isArray(profileView.followers) ? profileView.followers as RelationUser[] : []);
        setFollowing(Array.isArray(profileView.following) ? profileView.following as RelationUser[] : []);
        setUser((prev) => {
          if (!prev) return prev;
          return {
            ...prev,
            followersCount: profileView.counts.followers,
            followingCount: profileView.counts.following,
          };
        });
      }
    } catch (err: unknown) {
      console.warn('Failed to refresh relation lists:', err);
    } finally {
      setFollowDataLoading(false);
    }
  }, []);

  useEffect(() => {
    async function fetchProfile() {
      try {
        setIsLoading(true);
        
        const profileData = await getProfile();
        setCurrentUserId(profileData.id ?? null);

        // Get avatar URL from the avatar object
        const avatarPath = profileData.avatar?.file_path || profileData.avatar?.thumbnail_path;
        const avatarUrl = getAvatarUrl(avatarPath);

        // Map API response to user object
        setUser({
          avatarUrl,
          name: `${profileData.first_name} ${profileData.last_name}`,
          username: profileData.nickname || profileData.email.split('@')[0],
          bio: profileData.about_me || '',
          email: profileData.email,
          birthDate: profileData.date_of_birth,
          gender: profileData.gender || '',
          isPublic: !profileData.is_private,
          followersCount: profileData.followers_count || 0,
          followingCount: profileData.following_count || 0,
        });
        setIsPublic(!profileData.is_private);
      } catch (err: any) {
        const message = err?.message ?? 'Failed to load profile';
        
        // Fallback to mock data if enabled and API fails
        if (USE_MOCK_FALLBACK) {
          console.warn('API failed, using mock data:', message);
          setUser(mockUser);
          setIsPublic(mockUser.isPublic);
          setError(null); // Clear error since we're using fallback
        } else {
          // Original error handling
          if (message === 'Authentication required' || message.toLowerCase().includes('authentication')) {
            clearAuth();
            router.replace('/login');
            return;
          }
          setError(message);
        }
      } finally {
        setIsLoading(false);
      }
    }

    fetchProfile();
  }, [router]);

  useEffect(() => {
    if (!currentUserId) return;
    void refreshFollowData(currentUserId);
  }, [currentUserId, refreshFollowData]);

  useEffect(() => {
    if (activeTab !== 'Groups') return;
    let cancelled = false;
    setGroupsLoading(true);
    setGroupsError(null);
    getMyGroups()
      .then((list) => {
        if (!cancelled) setMyGroups(Array.isArray(list) ? list : []);
      })
      .catch((err) => {
        if (!cancelled) {
          setGroupsError(err?.message ?? 'Failed to load groups');
          setMyGroups([]);
        }
      })
      .finally(() => {
        if (!cancelled) setGroupsLoading(false);
      });
    return () => { cancelled = true; };
  }, [activeTab]);

  useEffect(() => {
    if (activeTab !== 'Posts' || !currentUserId) return;
    let cancelled = false;
    setPostsLoading(true);
    setPostsError(null);

    // Load posts from stable endpoint and enrich missing comment counts.
    getPostsByUserId(currentUserId)
      .then(async (basic) => {
        if (cancelled) return;

        const list = Array.isArray(basic) ? basic : [];
        const ids = list
          .map((p: any) => p?.id || p?.post_id)
          .filter((id: unknown): id is string => typeof id === 'string' && id.length > 0);

        const uniqueIds = Array.from(new Set(ids));
        const detailEntries = await Promise.all(
          uniqueIds.map(async (id) => {
            try {
              const detail = await getPostById(id);
              return [id, detail] as const;
            } catch {
              return [id, null] as const;
            }
          }),
        );

        if (cancelled) return;

        const detailsById: Record<string, any> = {};
        for (const [id, detail] of detailEntries) {
          detailsById[id] = detail;
        }

        const merged = list.map((post: any) => {
          const id = post?.id || post?.post_id;
          const detail = id ? detailsById[id] : null;
          const commentCount =
            typeof post?.comment_count === 'number'
              ? post.comment_count
              : typeof detail?.comment_count === 'number'
                ? detail.comment_count
                : Array.isArray(post?.comments)
                  ? post.comments.length
                  : 0;

          return {
            ...post,
            comment_count: commentCount,
          };
        });

        setMyPosts(merged);
      })
      .catch((err: any) => {
        if (cancelled) return;
        setPostsError(err?.message ?? 'Failed to load posts');
        setMyPosts([]);
      })
      .finally(() => {
        if (!cancelled) setPostsLoading(false);
      });
    return () => { cancelled = true; };
  }, [activeTab, currentUserId]);

  // Also pull the current feed once to reuse the exact same image URLs
  // the feed page sees (source of truth for post images).
  useEffect(() => {
    if (activeTab !== 'Posts') return;
    let cancelled = false;

    getFeed()
      .then((data: any) => {
        if (cancelled) return;
        const posts = Array.isArray(data) ? data : (data?.posts ?? []);
        const map: Record<string, string> = {};

        posts.forEach((post: any) => {
          const firstImage = post?.images?.[0];
          const url = firstImage?.thumbnail_url || firstImage?.url;
          if (post?.id && url) {
            map[post.id] = url;
          }
        });

        setFeedImagesByPostId(map);
      })
      .catch(() => {
        if (!cancelled) setFeedImagesByPostId({});
      });

    return () => {
      cancelled = true;
    };
  }, [activeTab]);

  function formatPostDate(isoDate: string): string {
    if (!isoDate) return '';
    try {
      const d = new Date(isoDate);
      return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
    } catch {
      return isoDate;
    }
  }

  const handleTogglePublic = async (newValue: boolean) => {
    const previousValue = isPublic;
    setIsPublic(newValue); // Optimistic update
    
    try {
      await updatePrivacy(!newValue); // API expects is_private (opposite of isPublic)
      console.log('Profile visibility changed to:', newValue ? 'public' : 'private');
    } catch (err: any) {
      console.error('Failed to update privacy:', err);
      setIsPublic(previousValue); // Revert on error
      // You could show a toast notification here
    }
  };

  const reloadFollowData = async () => {
    if (!currentUserId) return;
    await refreshFollowData(currentUserId);
  };

  const handleRemoveFollower = async (followerId: string) => {
    setFollowActionUserId(followerId);
    setFollowDataError(null);
    try {
      await removeFollower(followerId);
      await reloadFollowData();
    } catch (err: any) {
      setFollowDataError(err?.message ?? 'Failed to remove follower');
    } finally {
      setFollowActionUserId(null);
    }
  };

  const handleUnfollow = async (followeeId: string) => {
    setFollowActionUserId(followeeId);
    setFollowDataError(null);
    try {
      await unfollowUser(followeeId);
      await reloadFollowData();
    } catch (err: any) {
      setFollowDataError(err?.message ?? 'Failed to unfollow user');
    } finally {
      setFollowActionUserId(null);
    }
  };

  const handleFollowersClick = () => {
    setShowFollowersModal(true);
  };

  const handleFollowingClick = () => {
    setShowFollowingModal(true);
  };

  const handleAvatarUpload = async (file: File) => {
    setIsUploading(true);
    try {
      await uploadAvatar(file);
      // Refresh profile data to get updated avatar
      const profileData = await getProfile();
      const avatarPath = profileData.avatar?.file_path || profileData.avatar?.thumbnail_path;
      const avatarUrl = getAvatarUrl(avatarPath);
      
      setUser((prev) => (prev ? { ...prev, avatarUrl } : null));
    } catch (err: any) {
      console.error('Failed to upload avatar:', err);
      setError(err?.message || 'Failed to upload avatar');
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <main className="min-h-screen bg-parea-white px-16 py-12">
      <div className="max-w-7xl mx-auto">
        {/* Profile Card */}
        {user && (
          <ProfileWrap
            user={{
              ...user,
              isPublic, // Use local state for toggle
            }}
            isSelf={true}
            onTogglePublic={handleTogglePublic}
            onFollowersClick={handleFollowersClick}
            onFollowingClick={handleFollowingClick}
            onAvatarUpload={handleAvatarUpload}
          />
        )}

        {followDataError && (
          <p className="mt-4 text-regular text-parea-black">{followDataError}</p>
        )}

        {/* Placeholder for tabs and content below */}
        <div className="mt-8 max-w-312 w-full">
          <div className="flex items-center justify-between gap-6">
            <Tabs
              tabs={['Posts', 'Groups']}
              defaultTab="Posts"
              onTabChange={(tab) => setActiveTab(tab)}
            />
            {activeTab === 'Groups' && (
              <Button variant="primary" size="lg" onClick={() => setCreateGroupOpen(true)}>
                Create Group
              </Button>
            )}
          </div>

          {/* Tab content */}
          <div className="mt-6">
            {activeTab === 'Posts' && (
              <>
                {postsLoading && (
                  <p className="text-regular text-parea-black">Loading posts...</p>
                )}
                {!postsLoading && postsError && (
                  <p className="text-regular text-parea-black">{postsError}</p>
                )}
                {!postsLoading && !postsError && myPosts.length === 0 && (
                  <p className="text-regular text-parea-black">No posts yet.</p>
                )}
                {!postsLoading && !postsError && myPosts.length > 0 && (
                  <div className="flex flex-col items-start gap-0">
                    {myPosts.map((post, index) => {
                      const feedImage = feedImagesByPostId[post.id];
                      const rawImagePath = feedImage || post.thumbnail_url || post.image_url;
                      const imageUrl = getPostImageUrl(rawImagePath);
                      const commentCount =
                        typeof post.comment_count === 'number'
                          ? post.comment_count
                          : Array.isArray(post.comments)
                            ? post.comments.length
                            : 0;

                      const avatarSrc = user?.avatarUrl || '/user-avatar-default.png';
                      const avatarLabel = user?.username || post.nickname || 'Author';
                      const displayName = (user?.username || post.nickname || 'User').toUpperCase();

                      return (
                        <Card
                          key={post.id}
                          imageType="post"
                          imageSrc={imageUrl}
                          avatarSrc={avatarSrc}
                          avatarAlt={avatarLabel}
                          userName={displayName}
                          userDate={formatPostDate(post.created_at)}
                          title={post.title}
                          content={post.content}
                          href={`/post/${post.id}`}
                          imagePriority={index === 0}
                          commentCount={commentCount}
                        />
                      );
                    })}
                  </div>
                )}
              </>
            )}
            {activeTab === 'Groups' && (
              <>
                {groupsLoading && <p className="text-regular text-parea-black">Loading groups...</p>}
                {!groupsLoading && groupsError && <p className="text-regular text-parea-black">{groupsError}</p>}
                {!groupsLoading && !groupsError && myGroups.length === 0 && (
                  <p className="text-regular text-parea-black">No groups yet.</p>
                )}
                {!groupsLoading && !groupsError && myGroups.length > 0 && (
                  <div className="flex flex-col gap-4">
                    {myGroups.map((g) => (
                      <div
                        key={g.id}
                        onClick={() => router.push(`/group/${g.id}`)}
                        className="border border-parea-black p-6 bg-white cursor-pointer hover:shadow-[4px_4px_0_0_#000] transition-shadow"
                      >
                        <h3 className="text-2xl font-bold text-parea-black mb-2">
                          {g.title ?? g.name ?? 'Group'}
                        </h3>
                        {g.description && (
                          <p className="text-regular text-parea-black/70 mb-3">
                            {g.description}
                          </p>
                        )}
                        <div className="flex gap-4 text-sm text-parea-black/60">
                          <span>{g.member_count || 0} members</span>
                          <span>•</span>
                          <span>Created by {g.owner_nickname || 'Unknown'}</span>
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

      {createGroupOpen && (
        <CreateGroupModal
          isOpen={createGroupOpen}
          onClose={() => setCreateGroupOpen(false)}
          onSuccess={(groupId) => {
            setCreateGroupOpen(false);
            router.push(`/group/${groupId}`);
          }}
        />
      )}

      <FollowersModal
        isOpen={showFollowersModal}
        onClose={() => setShowFollowersModal(false)}
        heading="Followers"
        users={followers}
        onRemoveFollower={handleRemoveFollower}
        isActionLoading={followActionUserId}
      />

      <FollowersModal
        isOpen={showFollowingModal}
        onClose={() => setShowFollowingModal(false)}
        heading="Following"
        users={following}
        onUnfollow={handleUnfollow}
        isActionLoading={followActionUserId}
      />
    </main>
  );
}
