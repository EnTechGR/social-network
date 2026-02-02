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
import ProfileWrap from '@/components/ui/ProfileWrap';
import Tabs from '@/components/ui/Tabs';
import { getProfile } from '@/lib/api';
import { clearAuth } from '@/lib/auth';

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

export default function ProfilePage() {
  const router = useRouter();
  const [user, setUser] = useState(mockUser);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isPublic, setIsPublic] = useState(mockUser.isPublic);
  const [activeTab, setActiveTab] = useState('Posts');

  useEffect(() => {
    async function fetchProfile() {
      try {
        setIsLoading(true);
        const profileData = await getProfile();
        
        // Get avatar URL from the avatar object
        const avatarUrl = profileData.avatar?.file_path || profileData.avatar?.thumbnail_path || '/test-avatar.png';
        
        // Map API response to user object
        setUser({
          avatarUrl: avatarUrl.startsWith('http') ? avatarUrl : `http://localhost:8080${avatarUrl}`,
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
        if (message === 'Authentication required' || message.toLowerCase().includes('authentication')) {
          clearAuth();
          router.replace('/login');
          return;
        }
        setError(message);
      } finally {
        setIsLoading(false);
      }
    }

    fetchProfile();
  }, [router]);

  const handleTogglePublic = (newValue: boolean) => {
    setIsPublic(newValue);
    // TODO: Call API to update profile visibility
    console.log('Profile visibility changed to:', newValue ? 'public' : 'private');
  };

  const handleFollowersClick = () => {
    // TODO: Open followers modal/page
    console.log('Show followers list');
  };

  const handleFollowingClick = () => {
    // TODO: Open following modal/page
    console.log('Show following list');
  };

  return (
    <main className="min-h-screen bg-parea-white px-16 py-12">
      <div className="max-w-7xl mx-auto">
        {/* Profile Card */}
        <ProfileWrap
          user={{
            ...user,
            isPublic, // Use local state for toggle
          }}
          isSelf={true}
          onTogglePublic={handleTogglePublic}
          onFollowersClick={handleFollowersClick}
          onFollowingClick={handleFollowingClick}
        />

        {/* Placeholder for tabs and content below */}
        <div className="mt-8">
          <Tabs
            tabs={['Posts', 'Events', 'Reactions', 'Groups']}
            defaultTab="Posts"
            onTabChange={(tab) => setActiveTab(tab)}
          />

          {/* Tab content placeholder */}
          <div className="mt-6">
            {activeTab === 'Posts' && <p>Posts content...</p>}
            {activeTab === 'Events' && <p>Events content...</p>}
            {activeTab === 'Reactions' && <p>Reactions content...</p>}
            {activeTab === 'Groups' && <p>Groups content...</p>}
          </div>
        </div>
      </div>
    </main>
  );
}
