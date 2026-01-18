/**
 * app/group/page.tsx
 *
 * Group page - accessible at /group
 * Currently using mock data for testing the GroupWrap component.
 * TODO: Replace mock data with real API calls when backend is ready.
 */

'use client';

import { useState } from 'react';
import GroupWrap from '@/components/ui/GroupWrap';

// Mock group data for testing
const mockGroup = {
  imageUrl: '/group-book.jpg', // Reusing the test image for now
  name: 'Bookworms',
  description: 'Group for all bookworms around the world they can share ideas and best practices. Please repost comments and ideas.',
  admin: 'alexthegenius',
  createdDate: '05/11/2025',
  membersCount: 356,
};

export default function GroupPage() {
  const [isMember, setIsMember] = useState(false);

  const handleJoin = () => {
    console.log('Join request sent');
    // In real app, this would call API
  };

  const handleInvite = () => {
    console.log('Open invite modal');
  };

  const handleLeaveGroup = () => {
    setIsMember(false);
    console.log('Left group');
  };

  const handleMembersClick = () => {
    console.log('Show members list');
  };

  return (
    <main className="min-h-screen bg-parea-white px-16 py-12">
      <div className="max-w-[1280px] mx-auto space-y-8">
        {/* Toggle for testing */}
        <div className="flex gap-4 items-center">
          <span className="font-medium">Test Mode:</span>
          <button
            onClick={() => setIsMember(!isMember)}
            className="px-4 py-2 bg-parea-yellow border border-parea-black rounded"
          >
            {isMember ? 'Switch to Non-Member View' : 'Switch to Member View'}
          </button>
        </div>

        {/* Group Card */}
        <GroupWrap
          group={mockGroup}
          isMember={isMember}
          onJoin={handleJoin}
          onInvite={handleInvite}
          onLeaveGroup={handleLeaveGroup}
          onMembersClick={handleMembersClick}
        />
      </div>
    </main>
  );
}
