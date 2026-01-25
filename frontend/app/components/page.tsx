'use client';

import { useState } from 'react';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Avatar from '@/components/ui/Avatar';
import Tabs from '@/components/ui/Tabs';
import ToggleButton from '@/components/ui/ToggleButton';
import IconButton from '@/components/ui/IconButtons';
import CreatePostModal from '@/components/ui/CreatePostModal';
import CreateEventModal from '@/components/ui/CreateEventModal';
import FollowersModal from '@/components/ui/FollowersModal';
import { Card } from '@/components/ui/Card';
import ProfileWrap from '@/components/ui/ProfileWrap';
import GroupWrap from '@/components/ui/GroupWrap';
import SidebarToggle from '@/components/ui/SidebarToggle';
import SearchSuggestions from '@/components/ui/SearchSuggestions';
import { Navbar } from '@/components/ui/Navbar';
import Header from '@/components/Header';
import AuthLayout from '@/components/auth/AuthLayout';

export default function ComponentsPage() {
  const [inputValue, setInputValue] = useState('');
  const [toggleOn, setToggleOn] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className="min-h-screen bg-parea-white p-8">
      <h1 className="text-h1 mb-12">Component Library</h1>

      {/* Buttons */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Buttons</h2>
        <div className="flex flex-wrap gap-4 items-center">
          <Button variant="primary" size="lg">Primary Large</Button>
          <Button variant="primary" size="md">Primary Medium</Button>
          <Button variant="primary" size="sm">Primary Small</Button>
          <Button variant="secondary" size="lg">Secondary</Button>
          <div className="flex flex-col gap-4">
            <Button variant="tertiary" size="lg">Regular Tertiary</Button>
            <Button
              variant="tertiary"
              size="lg"
              inactiveText="FOLLOWING"
              activeText="FOLLOW"
            />
          </div>
          <Button variant="primary" size="lg" disabled>Disabled</Button>
        </div>
      </section>

      {/* Icon Buttons */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Icon Buttons</h2>
        <div className="flex flex-wrap gap-4 items-center">
          <IconButton variant="arrow-left" aria-label="Go back" />
          <IconButton variant="arrow-right" aria-label="Go forward" />
          <IconButton variant="close" aria-label="Close" />
          <IconButton variant="arrow-left" aria-label="Back" text="BACK" />
          <IconButton variant="arrow-right" aria-label="Next" text="NEXT" />
          <IconButton variant="close" aria-label="Close" text="CLOSE" />
        </div>
      </section>

      {/* Input */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Input</h2>
        <div className="max-w-md space-y-4">
          <Input
            value={inputValue}
            onChange={setInputValue}
            placeholder="PLACEHOLDER TEXT"
            label="Label"
          />
          <Input
            value=""
            onChange={() => { }}
            placeholder="WITH ERROR"
            error="This field is required"
          />
          <Input
            value=""
            onChange={() => { }}
            placeholder="DISABLED"
            disabled
          />
        </div>
      </section>

      {/* Avatar */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Avatar</h2>
        <div className="flex flex-wrap gap-6 items-end">
          <div className="text-center">
            <Avatar size="sm" type="user" />
            <p className="mt-2 text-sm">SM User</p>
          </div>
          <div className="text-center">
            <Avatar size="md" type="user" />
            <p className="mt-2 text-sm">MD User</p>
          </div>
          <div className="text-center">
            <Avatar size="lg" type="user" />
            <p className="mt-2 text-sm">LG User</p>
          </div>
          <div className="text-center">
            <Avatar size="md" type="group" />
            <p className="mt-2 text-sm">Group</p>
          </div>
          <div className="text-center">
            <Avatar size="md" type="user" isSelf />
            <p className="mt-2 text-sm">Self (yellow shadow)</p>
          </div>
          <div className="text-center">
            <Avatar size="md" type="user" isSelf={false} />
            <p className="mt-2 text-sm">Not Self (black shadow)</p>
          </div>
        </div>
      </section>

      {/* Tabs */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Tabs</h2>
        <Tabs tabs={['POSTS', 'EVENTS', 'USERS', 'GROUPS']} />
      </section>

      {/* Toggle Button */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Toggle Button</h2>
        <div className="flex gap-6 items-center">
          <div className="flex items-center gap-2">
            <ToggleButton
              isOn={toggleOn}
              onChange={setToggleOn}
              aria-label="Toggle setting"
            />
            <span className="text-sm">{toggleOn ? 'ON' : 'OFF'}</span>
          </div>
          <div className="flex items-center gap-2">
            <ToggleButton
              isOn={true}
              onChange={() => { }}
              aria-label="Always on"
            />
            <span className="text-sm">Always ON</span>
          </div>
          <div className="flex items-center gap-2">
            <ToggleButton
              isOn={false}
              onChange={() => { }}
              aria-label="Always off"
            />
            <span className="text-sm">Always OFF</span>
          </div>
          <div className="flex items-center gap-2">
            <ToggleButton
              isOn={false}
              onChange={() => { }}
              disabled
              aria-label="Disabled"
            />
            <span className="text-sm">Disabled</span>
          </div>
        </div>
      </section>

      {/* Modal */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Create Post Modal</h2>
        <CreatePostModal isOpen={false} onClose={() => { }} preview />
      </section>

      {/* Create Event Modal */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Create Event Modal</h2>
        <CreateEventModal isOpen={false} onClose={() => { }} preview />
      </section>

      {/* Followers Modal */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Followers Modal</h2>
        <FollowersModal isOpen={false} onClose={() => { }} heading="Followers" preview />
      </section>

      {/* Profile Wrap */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Profile Wrap</h2>
        <div className="space-y-8">
          <div>
            <p className="text-small mb-4">Own Profile (with toggle)</p>
            <ProfileWrap
              user={{
                avatarUrl: '/test-avatar.png',
                name: 'John Doe',
                username: 'johndoe',
                bio: 'Software developer passionate about building great user experiences. Love hiking and photography on weekends.',
                email: 'john@example.com',
                birthDate: 'Jan 15, 1990',
                isPublic: true,
                followersCount: 1234,
                followingCount: 567,
              }}
              isSelf={true}
              onTogglePublic={() => { }}
              onFollowersClick={() => { }}
              onFollowingClick={() => { }}
            />
          </div>
          <div>
            <p className="text-small mb-4">Other User Profile (no toggle)</p>
            <ProfileWrap
              user={{
                name: 'Jane Smith',
                username: 'janesmith',
                bio: 'Designer and creative thinker.',
                email: 'jane@example.com',
                birthDate: 'Mar 22, 1995',
                isPublic: true,
                followersCount: 890,
                followingCount: 234,
              }}
              isSelf={false}
              onFollowersClick={() => { }}
              onFollowingClick={() => { }}
            />
          </div>
        </div>
      </section>

      {/* Group Wrap */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Group Wrap</h2>
        <div className="space-y-8">
          <div>
            <p className="text-small mb-4">Member View (Invite + Leave buttons)</p>
            <GroupWrap
              group={{
                imageUrl: '/test-avatar.png',
                name: 'Photography Enthusiasts',
                description: 'A community for photographers of all skill levels to share their work, get feedback, and learn from each other.',
                admin: 'John Doe',
                createdDate: 'Jan 10, 2025',
                membersCount: 156,
              }}
              isMember={true}
              onInvite={() => { }}
              onLeaveGroup={() => { }}
              onMembersClick={() => { }}
            />
          </div>
          <div>
            <p className="text-small mb-4">Non-Member View (Join button)</p>
            <GroupWrap
              group={{
                name: 'Hiking Adventures',
                description: 'Explore trails and nature with fellow hikers.',
                admin: 'Jane Smith',
                createdDate: 'Mar 5, 2025',
                membersCount: 89,
              }}
              isMember={false}
              onJoin={() => { }}
              onMembersClick={() => { }}
            />
          </div>
          <div>
            <p className="text-small mb-4">Pending Join Request</p>
            <GroupWrap
              group={{
                name: 'Private Book Club',
                description: 'Monthly book discussions and reading recommendations.',
                admin: 'Alex Johnson',
                createdDate: 'Feb 20, 2025',
                membersCount: 42,
              }}
              isMember={false}
              isPending={true}
              onMembersClick={() => { }}
            />
          </div>
        </div>
      </section>

      {/* Sidebar Toggle */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Sidebar Toggle</h2>
        <div className="flex gap-6 items-center">
          <div className="flex items-center gap-4 p-4 bg-parea-black rounded-lg">
            <SidebarToggle isOpen={sidebarOpen} onClick={() => setSidebarOpen(!sidebarOpen)} />
            <span className="text-parea-white text-sm">{sidebarOpen ? 'Open' : 'Closed'}</span>
          </div>
          <div className="flex items-center gap-4 p-4 bg-parea-black rounded-lg">
            <SidebarToggle isOpen={true} onClick={() => { }} />
            <span className="text-parea-white text-sm">Always Open</span>
          </div>
          <div className="flex items-center gap-4 p-4 bg-parea-black rounded-lg">
            <SidebarToggle isOpen={false} onClick={() => { }} />
            <span className="text-parea-white text-sm">Always Closed</span>
          </div>
        </div>
      </section>

      {/* Search Suggestions */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Search Suggestions</h2>
        <div className="relative border border-parea-border rounded-lg overflow-hidden">
          <SearchSuggestions />
        </div>
      </section>

      {/* Navbar */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Navbar</h2>
        <div className="border border-parea-border rounded-lg overflow-hidden">
          <Navbar />
        </div>
      </section>

      {/* Header (Legacy) */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Header (Legacy)</h2>
        <p className="text-small mb-4 text-parea-black/60">Note: This is an older header component with different styling.</p>
        <div className="border border-parea-border rounded-lg overflow-hidden">
          <Header />
        </div>
      </section>

      {/* Auth Layout */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Auth Layout</h2>
        <p className="text-small mb-4 text-parea-black/60">Split-screen layout used for login/signup pages.</p>
        <div className="border border-parea-border rounded-lg overflow-hidden h-100">
          <AuthLayout>
            <div className="p-8 text-center">
              <h3 className="text-h4 mb-4">Form Content Goes Here</h3>
              <p className="text-regular text-parea-black/60">Login or signup form would be placed in this area.</p>
            </div>
          </AuthLayout>
        </div>
      </section>

      {/* Typography */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Typography</h2>
        <div className="space-y-4">
          <h1 className="text-h1">Heading 1</h1>
          <h2 className="text-h2">Heading 2</h2>
          <h3 className="text-h3">Heading 3</h3>
          <h4 className="text-h4">Heading 4</h4>
          <h5 className="text-h5">Heading 5</h5>
          <h6 className="text-h6">Heading 6</h6>
          <p className="text-regular">Regular body text (16px)</p>
          <p className="text-small">Small text (14px)</p>
          <p className="text-tiny">Tiny text (12px)</p>
          <p className="label">LABEL TEXT (IBM Plex Mono, uppercase)</p>
        </div>
      </section>

      {/* Colors */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Colors</h2>
        <div className="flex flex-wrap gap-4">
          <div className="w-24 h-24 bg-parea-black border border-parea-black flex items-end p-2">
            <span className="text-tiny text-white">Black</span>
          </div>
          <div className="w-24 h-24 bg-parea-white border border-parea-black flex items-end p-2">
            <span className="text-tiny">White</span>
          </div>
          <div className="w-24 h-24 bg-parea-grey border border-parea-black flex items-end p-2">
            <span className="text-tiny">Grey</span>
          </div>
          <div className="w-24 h-24 bg-parea-yellow border border-parea-black flex items-end p-2">
            <span className="text-tiny">Yellow</span>
          </div>
        </div>
      </section>

      {/* Card */}
      <section className="mb-12">
        <h2 className="text-h4 mb-6 border-b border-parea-black pb-2">Card</h2>
        <div className="space-y-4">
          <Card
            imageType="post"
            imageSrc="/postCardImage.png"
            avatarSrc="/test-avatar.png"
            avatarAlt="Test user"
            userName="John Doe"
            userDate="Jan 20, 2026"
          />
          <Card
            imageType="event"
            imageSrc="/eventCardImage.png"
            avatarSrc="/test-avatar.png"
            avatarAlt="Event creator"
            userName="Jane Smith"
            userDate="Jan 21, 2026"
          />
          <Card
            imageType="post"
            avatarSrc="/test-avatar.png"
            avatarAlt="User"
            userName="Default Post"
            userDate="Jan 22, 2026"
          />
          <Card
            imageType="event"
            avatarSrc="/test-avatar.png"
            avatarAlt="User"
            userName="Default Event"
            userDate="Jan 23, 2026"
          />
        </div>
      </section>
    </div>
  );
}
