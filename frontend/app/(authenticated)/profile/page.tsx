/**
 * app/profile/page.tsx
 *
 * Profile page - accessible at /profile
 * Currently using mock data for testing the ProfileCard component.
 * TODO: Replace mock data with real API calls when backend is ready.
 */

'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import ProfileWrap from '@/components/ui/ProfileWrap';
import Tabs from '@/components/ui/Tabs';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import CreateGroupModal from '@/components/ui/CreateGroupModal';
import { getProfile, updatePrivacy, uploadAvatar, getAvatarUrl, getPostsByUserId, getMyGroups } from '@/lib/api';
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
  isPublic: true,
  followersCount: 356,
  followingCount: 250,
};

type ProfileUser = typeof mockUser;

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

  const handleFollowersClick = () => {
    // TODO: Open followers modal/page
    console.log('Show followers list');
  };

  const handleFollowingClick = () => {
    // TODO: Open following modal/page
    console.log('Show following list');
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
                    {myPosts.map((post, index) => (
                      <Card
                        key={post.id}
                        imageType="post"
                        imageSrc={post.image_url || post.thumbnail_url}
                        avatarSrc="/user-avatar-default.png"
                        avatarAlt={post.nickname ?? 'Author'}
                        userName={(post.nickname ?? 'User').toUpperCase()}
                        userDate={formatPostDate(post.created_at)}
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
    </main>
  );
}
