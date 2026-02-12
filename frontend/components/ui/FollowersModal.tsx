'use client';

import { useState, useRef } from 'react';
import IconButton from './IconButtons';
import { Search } from 'lucide-react';
import Button from './Button';
import Image from 'next/image';

interface FollowersModalProps {
  isOpen: boolean;
  onClose: () => void;
  heading?: 'Followers' | 'Following' | 'Members';
  preview?: boolean;
}

export default function FollowersModal({ isOpen, onClose, heading = 'Followers', preview = false }: FollowersModalProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const searchInputRef = useRef<HTMLInputElement>(null);
  
  // State for follow/following toggle (only needed when heading === 'Following')
  const [followingStates, setFollowingStates] = useState<Record<number, boolean>>({
    1: true,
    2: true,
    3: true,
    4: true,
    5: true,
    6: true,
    7: true,
    8: true,
    9: true,
    10: true,
  });

  const mockUsers = [
    { id: 1, name: 'John Doe' },
    { id: 2, name: 'Jane Smith' },
    { id: 3, name: 'Bob Johnson' },
    { id: 4, name: 'Alice Brown' },
    { id: 5, name: 'Charlie Davis' },
    { id: 6, name: 'Diana Wilson' },
    { id: 7, name: 'Ethan Martinez' },
    { id: 8, name: 'Fiona Garcia' },
    { id: 9, name: 'George Taylor' },
    { id: 10, name: 'Hannah Anderson' },
  ];

  if (!isOpen && !preview) return null;

  const handleFollowToggle = (userId: number) => {
    setFollowingStates((prev) => ({
      ...prev,
      [userId]: !prev[userId],
    }));
    // TODO: Call API to follow/unfollow user
  };

  const modalContent = (
    <div className="relative w-full max-w-145 bg-white border border-parea-black shadow-[8px_8px_0_0_#000]">
      {/* Header */}
      <div className="relative h-17 border-b border-parea-black overflow-hidden bg-parea-white">
        <Image
          src="/modal-header-pattern.svg"
          alt=""
          fill
          className="object-cover"
        />
        <IconButton
          variant="close"
          onClick={onClose}
          aria-label="Close modal"
          className="absolute top-4 right-8 z-10"
        />
      </div>

      {/* Content */}
      <div
        className="
            flex
            flex-col
            items-start
            gap-8
            self-stretch
            p-8
            overflow-hidden
        "
      >
        {/* Heading - displays the prop value: "Followers", "Following", or "Members" */}
        <h4
          className="
            text-[32px]
            font-bold
            leading-[130%]
            text-black
          "
        >
          {heading}
        </h4>

        {/* Container */}
        <div
          className="
            flex
            flex-col
            items-start
            gap-8
            self-stretch
          "
        >
          {/* Search Container */}
          <div
            className="
              flex
              flex-col
              items-start
              gap-2
              self-stretch
            "
          >
            <div
              className="
                w-full
                rounded-full
                border
                border-parea-border
                flex
                flex-row
                items-center
                px-3
                py-2
                gap-3
                bg-parea-white
              "
            >
              <Search className="w-6 h-6 text-parea-black/70 flex-shrink-0" />
              <input
                ref={searchInputRef}
                type="text"
                placeholder="Search..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="
                  flex-1
                  border-none
                  outline-none
                  bg-transparent
                  text-black/60
                  font-mono
                  text-[15px]
                  font-medium
                  leading-[1.5]
                  tracking-[-0.15px]
                  uppercase
                  placeholder:text-black/60
                  w-full
                "
              />
            </div>
          </div>

          {/* List Container */}
          <div
            className="
              flex
              flex-col
              items-start
              gap-2
              self-stretch
              relative
              max-h-[400px]
              overflow-y-auto
              overflow-x-hidden
              pb-32
            "
          >
            {mockUsers.map((user) => (
              <div
                key={user.id}
                className="
                  flex
                  h-14
                  p-2
                  items-center
                  gap-2
                  self-stretch
                "
              >
                {/* Avatar Image */}
                <div
                  className="
                    w-12
                    h-12
                    rounded-full
                    flex-shrink-0
                    aspect-square
                    bg-parea-grey
                    bg-center
                    bg-cover
                    bg-no-repeat
                  "
                  style={{
                    backgroundImage: 'url(/test-avatar.png)',
                  }}
                />

                {/* Profile Info */}
                <div
                  className="
                    flex
                    flex-col
                    items-start
                    flex-1
                  "
                >
                  <p
                    className="
                      text-black
                      font-mono
                      text-[15px]
                      font-medium
                      uppercase
                      tracking-[-0.15px]
                    "
                  >
                    {user.name}
                  </p>
                </div>

                {/* Action Button - Conditional rendering based on heading prop */}
                {heading === 'Followers' ? (
                  // Case 1: Followers modal - Show "Remove" button
                  <Button
                    variant="tertiary"
                    onClick={() => {
                      // TODO: Handle remove follower action
                      console.log('Remove follower:', user.id);
                    }}
                  >
                    Remove
                  </Button>
                ) : heading === 'Following' ? (
                  // Case 2: Following modal - Show follow/following toggle button
                  <Button
                    variant="tertiary"
                    isActive={followingStates[user.id]}
                    onActiveChange={() => handleFollowToggle(user.id)}
                    activeText="FOLLOW"
                    inactiveText="FOLLOWING"
                    onClick={() => {
                      // onClick is handled by onActiveChange
                    }}
                  />
                ) : (
                  // Case 3: Members modal - No button (null)
                  null
                )}
              </div>
            ))}
          </div>

          {/* Fade Effect Overlay */}
          <div
            className="
              absolute
              w-[767px]
              h-[141px]
              rounded-[767px]
              bg-parea-white
              blur-[30px]
              pointer-events-none
            "
            style={{
              left: '-128px',
              bottom: '-32px',
            }}
          />
        </div>
      </div>
    </div>
  );

  if (preview) {
    return modalContent;
  }

  return (
    <div
      className="
        fixed
        inset-0
        z-50
        flex
        items-center
        justify-center
      "
      onClick={onClose}
    >
      <div
        className="
          absolute
          inset-0
          bg-black/50
        "
      />
      <div onClick={(e) => e.stopPropagation()}>
        {modalContent}
      </div>
    </div>
  );
}