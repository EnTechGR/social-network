/**
 * app/profile/page.tsx
 *
 * Profile page - accessible at /profile
 * Currently using mock data for testing the ProfileCard component.
 * TODO: Replace mock data with real API calls when backend is ready.
 */

'use client';

import { useState } from 'react';
import ProfileCard from '@/components/ui/ProfileCard';

// Mock user data for testing - replace with API call later
const mockUser = {
  avatarUrl: '/test-avatar.jpg',
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
  // Local state to handle toggle (will be replaced with API call)
  const [isPublic, setIsPublic] = useState(mockUser.isPublic);

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
    <main className="min-h-screen bg-parea-white p-8">
      <div className="max-w-[1280px] mx-auto">
        {/* Profile Card */}
        <ProfileCard
          user={{
            ...mockUser,
            isPublic, // Use local state for toggle
          }}
          isSelf={true}
          onTogglePublic={handleTogglePublic}
          onFollowersClick={handleFollowersClick}
          onFollowingClick={handleFollowingClick}
        />

        {/* Placeholder for tabs and content below */}
        <div className="mt-8 p-8 border border-dashed border-parea-border rounded text-center text-parea-black/50">
          <p className="label">Tabs component will go here</p>
          <p className="text-small mt-2">(Posts, Events, etc.)</p>
        </div>
      </div>
    </main>
  );
}
