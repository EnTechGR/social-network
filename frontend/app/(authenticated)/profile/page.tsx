/**
 * app/profile/page.tsx
 *
 * Profile page - accessible at /profile
 * Currently using mock data for testing the ProfileCard component.
 * TODO: Replace mock data with real API calls when backend is ready.
 */

'use client';

import { useState, useEffect } from 'react';
import ProfileWrap from '@/components/ui/ProfileWrap';
import Tabs from '@/components/ui/Tabs';
import { getProfile } from '@/lib/api';

export default function ProfilePage() {
  const [user, setUser] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isPublic, setIsPublic] = useState(true);
  const [activeTab, setActiveTab] = useState('Posts');

  useEffect(() => {
    async function fetchProfile() {
      try {
        setIsLoading(true);
        const profileData = await getProfile();
        
        // Get avatar URL from the avatar object
        const avatarPath = profileData.avatar?.file_path || profileData.avatar?.thumbnail_path;
        
        // Construct proper avatar URL
        let avatarUrl = '/user-avatar-default.png'; // default fallback
        if (avatarPath) {
          // Backend serves files at /static/ (which maps to ./uploads directory)
          // If path is like "uploads/file.png", we need "/static/file.png"
          const filename = avatarPath.replace(/^uploads\//, '');
          avatarUrl = `http://localhost:8080/static/${filename}`;
        }
        
        // Format date of birth to only show date (remove time)
        const formatDate = (dateString: string) => {
          const date = new Date(dateString);
          return date.toLocaleDateString('en-GB', { 
            day: '2-digit', 
            month: '2-digit', 
            year: 'numeric' 
          });
        };
        
        // Map API response to user object
        setUser({
          avatarUrl,
          name: `${profileData.first_name} ${profileData.last_name}`,
          username: profileData.nickname,
          bio: profileData.about_me || '',
          email: profileData.email,
          birthDate: formatDate(profileData.date_of_birth),
          isPublic: !profileData.is_private,
          followersCount: profileData.followers_count || 0,
          followingCount: profileData.following_count || 0,
        });
        setIsPublic(!profileData.is_private);
      } catch (err: any) {
        console.error('Failed to fetch profile:', err);
        setError(err.message || 'Failed to load profile');
      } finally {
        setIsLoading(false);
      }
    }

    fetchProfile();
  }, []);

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
