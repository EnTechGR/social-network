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
import Link from 'next/link';
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
  getMyGroups,
  getUserProfile,
  getFollowRequests,
  acceptFollowRequest,
  declineFollowRequest,
  removeFollower,
  unfollowUser,
  getForumUsers,
  type FollowRelationship,
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
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isPublic, setIsPublic] = useState(true);
  const [activeTab, setActiveTab] = useState('Posts');
  const [isUploading, setIsUploading] = useState(false);
  const [myPosts, setMyPosts] = useState<any[]>([]);
  const [postsLoading, setPostsLoading] = useState(false);
  const [postsError, setPostsError] = useState<string | null>(null);
  const [myGroups, setMyGroups] = useState<any[]>([]);
  const [groupsLoading, setGroupsLoading] = useState(false);
  const [groupsError, setGroupsError] = useState<string | null>(null);
  const [createGroupOpen, setCreateGroupOpen] = useState(false);
  const [followers, setFollowers] = useState<RelationUser[]>([]);
  const [following, setFollowing] = useState<RelationUser[]>([]);
  const [pendingRequests, setPendingRequests] = useState<FollowRelationship[]>([]);
  const [followDataLoading, setFollowDataLoading] = useState(false);
  const [followDataError, setFollowDataError] = useState<string | null>(null);
  const [followActionUserId, setFollowActionUserId] = useState<string | null>(null);
  const [showFollowersList, setShowFollowersList] = useState(false);
  const [showFollowingList, setShowFollowingList] = useState(false);
  const [showFollowersModal, setShowFollowersModal] = useState(false);
  const [showFollowingModal, setShowFollowingModal] = useState(false);
  const [directoryById, setDirectoryById] = useState<Record<string, RelationUser>>({});

  const formatRelationName = (entry: RelationUser | undefined): string => {
    if (!entry) return 'Unknown user';
    const fullName = [entry.first_name, entry.last_name].filter(Boolean).join(' ').trim();
    return fullName || entry.nickname || entry.email || entry.user_id;
  };

  const getUserLabel = (userId: string): string => {
    const fromDirectory = directoryById[userId];
    if (fromDirectory) return formatRelationName(fromDirectory);

    const fromFollowers = followers.find((item) => item.user_id === userId);
    if (fromFollowers) return formatRelationName(fromFollowers);

    const fromFollowing = following.find((item) => item.user_id === userId);
    if (fromFollowing) return formatRelationName(fromFollowing);

    return userId;
  };

  const refreshFollowData = useCallback(async (userId: string) => {
    setFollowDataLoading(true);
    setFollowDataError(null);

    try {
      const [profileView, requests, users] = await Promise.all([
        getUserProfile(userId),
        getFollowRequests(),
        getForumUsers(),
      ]);

      if (!profileView.privateProfile) {
        const followerList = Array.isArray(profileView.followers) ? profileView.followers as RelationUser[] : [];
        const followingList = Array.isArray(profileView.following) ? profileView.following as RelationUser[] : [];

        setFollowers(followerList);
        setFollowing(followingList);

        setUser((prev) => {
          if (!prev) return prev;
          return {
            ...prev,
            followersCount: profileView.counts.followers,
            followingCount: profileView.counts.following,
          };
        });
      }

      setPendingRequests(Array.isArray(requests.pending) ? requests.pending : []);

      const byId = (users || []).reduce<Record<string, RelationUser>>((acc, item) => {
        acc[item.id] = {
          user_id: item.id,
          nickname: item.nickname,
          first_name: item.first_name,
          last_name: item.last_name,
          email: item.email,
        };
        return acc;
      }, {});
      setDirectoryById(byId);
    } catch (err: any) {
      setFollowDataError(err?.message ?? 'Failed to load follower data');
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
    getPostsByUserId(currentUserId)
      .then((data) => {
        if (!cancelled) setMyPosts(Array.isArray(data) ? data : []);
      })
      .catch((err) => {
        if (!cancelled) {
          setPostsError(err?.message ?? 'Failed to load posts');
          setMyPosts([]);
        }
      })
      .finally(() => {
        if (!cancelled) setPostsLoading(false);
      });
    return () => { cancelled = true; };
  }, [activeTab, currentUserId]);

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

  const handleAcceptRequest = async (followerId: string) => {
    setFollowActionUserId(followerId);
    setFollowDataError(null);
    try {
      await acceptFollowRequest(followerId);
      await reloadFollowData();
    } catch (err: any) {
      setFollowDataError(err?.message ?? 'Failed to accept follow request');
    } finally {
      setFollowActionUserId(null);
    }
  };

  const handleDeclineRequest = async (followerId: string) => {
    setFollowActionUserId(followerId);
    setFollowDataError(null);
    try {
      await declineFollowRequest(followerId);
      await reloadFollowData();
    } catch (err: any) {
      setFollowDataError(err?.message ?? 'Failed to decline follow request');
    } finally {
      setFollowActionUserId(null);
    }
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

        <section className="mt-6 rounded border border-parea-black bg-parea-white p-4">
          <h2 className="text-small font-medium uppercase text-parea-black">Pending Follow Requests</h2>
          {followDataLoading ? (
            <p className="mt-3 text-regular text-parea-black">Loading follow requests...</p>
          ) : pendingRequests.length === 0 ? (
            <p className="mt-3 text-regular text-parea-black">No pending follow requests.</p>
          ) : (
            <ul className="mt-3 flex flex-col gap-3">
              {pendingRequests.map((request) => (
                <li key={`${request.follower_id}:${request.followee_id}`} className="flex flex-wrap items-center gap-2">
                  <span className="text-regular text-parea-black">{getUserLabel(request.follower_id)}</span>
                  <button
                    type="button"
                    onClick={() => { void handleAcceptRequest(request.follower_id); }}
                    disabled={followActionUserId === request.follower_id}
                    className="rounded border border-parea-black bg-parea-yellow px-3 py-1 text-small font-medium uppercase text-parea-black hover:opacity-90 disabled:opacity-60 disabled:cursor-not-allowed"
                  >
                    Accept
                  </button>
                  <button
                    type="button"
                    onClick={() => { void handleDeclineRequest(request.follower_id); }}
                    disabled={followActionUserId === request.follower_id}
                    className="rounded border border-parea-black bg-parea-white px-3 py-1 text-small font-medium uppercase text-parea-black hover:opacity-90 disabled:opacity-60 disabled:cursor-not-allowed"
                  >
                    Decline
                  </button>
                </li>
              ))}
            </ul>
          )}
        </section>

        {/* Placeholder for tabs and content below */}
        <div className="mt-8">
          <Tabs
            tabs={['Posts', 'Events', 'Reactions', 'Groups']}
            defaultTab="Posts"
            onTabChange={(tab) => setActiveTab(tab)}
          />

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
                      const rawImagePath = post.thumbnail_url || post.image_url;
                      const imageUrl = getPostImageUrl(rawImagePath);

                      return (
                        <Card
                          key={post.id}
                          imageType="post"
                          imageSrc={imageUrl}
                          avatarSrc="/user-avatar-default.png"
                          avatarAlt={post.nickname ?? 'Author'}
                          userName={(post.nickname ?? 'User').toUpperCase()}
                          userDate={formatPostDate(post.created_at)}
                          title={post.title}
                          content={post.content}
                          href={`/post/${post.id}`}
                          imagePriority={index === 0}
                        />
                      );
                    })}
                  </div>
                )}
              </>
            )}
            {activeTab === 'Events' && <p className="text-regular text-parea-black">Events content...</p>}
            {activeTab === 'Reactions' && <p className="text-regular text-parea-black">Reactions content...</p>}
            {activeTab === 'Groups' && (
              <>
                <div className="mb-4">
                  <Button variant="primary" size="lg" onClick={() => setCreateGroupOpen(true)}>
                    Create Group
                  </Button>
                </div>
                {groupsLoading && <p className="text-regular text-parea-black">Loading groups...</p>}
                {!groupsLoading && groupsError && <p className="text-regular text-parea-black">{groupsError}</p>}
                {!groupsLoading && !groupsError && myGroups.length === 0 && (
                  <p className="text-regular text-parea-black">No groups yet.</p>
                )}
                {!groupsLoading && !groupsError && myGroups.length > 0 && (
                  <ul className="flex flex-col gap-2">
                    {myGroups.map((g) => (
                      <li key={g.id}>
                        <Link href={`/group/${g.id}`} className="text-regular text-parea-black underline hover:no-underline">
                          {g.title ?? g.name ?? 'Group'}
                        </Link>
                      </li>
                    ))}
                  </ul>
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
