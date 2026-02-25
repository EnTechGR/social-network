'use client';

import { useState, useRef, useMemo } from 'react';
import IconButton from './IconButtons';
import { Search } from 'lucide-react';
import Button from './Button';
import Image from 'next/image';
import { getAvatarUrl } from '@/lib/api';

export interface FollowerUser {
  user_id: string;
  nickname?: string;
  first_name?: string;
  last_name?: string;
  email?: string;
  avatar?: {
    file_path?: string;
    thumbnail_path?: string;
  };
}

interface FollowersModalProps {
  isOpen: boolean;
  onClose: () => void;
  heading?: 'Followers' | 'Following' | 'Members';
  preview?: boolean;
  users?: FollowerUser[];
  onRemoveFollower?: (userId: string) => Promise<void>;
  onUnfollow?: (userId: string) => Promise<void>;
  onInvite?: (userId: string) => Promise<void>;
  isActionLoading?: string | null;
}

export default function FollowersModal({
  isOpen,
  onClose,
  heading = 'Followers',
  preview = false,
  users = [],
  onRemoveFollower,
  onUnfollow,
  onInvite,
  isActionLoading = null,
}: FollowersModalProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const searchInputRef = useRef<HTMLInputElement>(null);

  const formatUserName = (user: FollowerUser): string => {
    const fullName = [user.first_name, user.last_name].filter(Boolean).join(' ').trim();
    return fullName || user.nickname || user.email || user.user_id;
  };

  const filteredUsers = useMemo(() => {
    if (!searchQuery.trim()) return users;
    const query = searchQuery.toLowerCase();
    return users.filter((user) => {
      const name = formatUserName(user).toLowerCase();
      return name.includes(query);
    });
  }, [users, searchQuery]);

  const handleAction = async (userId: string) => {
    if (heading === 'Followers' && onRemoveFollower) {
      await onRemoveFollower(userId);
    } else if (heading === 'Following' && onUnfollow) {
      await onUnfollow(userId);
    } else if (heading === 'Members' && onInvite) {
      await onInvite(userId);
    }
  };

  if (!isOpen && !preview) return null;

  const modalContent = (
    <div className="relative w-145 h-172 flex flex-col bg-white border border-parea-black shadow-[8px_8px_0_0_#000]">
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
            flex-1
            min-h-0
        "
      >
        {/* Heading - displays the prop value: "Followers", "Following", or "Members" */}
        <h4
          className="
            text-h4
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
            flex-1
            min-h-0
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
              <Search className="w-6 h-6 text-parea-black/70 shrink-0" />
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
                  leading-relaxed
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
              flex-1
              min-h-0
              overflow-y-auto
              overflow-x-hidden
            "
          >
            {filteredUsers.length === 0 ? (
              <p className="text-regular text-parea-black py-4">
                {searchQuery ? 'No users found.' : `No ${heading.toLowerCase()} yet.`}
              </p>
            ) : (
              filteredUsers.map((user) => {
                const avatarUrl = getAvatarUrl(user.avatar?.thumbnail_path || user.avatar?.file_path);

                return (
                  <div
                    key={user.user_id}
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
                        shrink-0
                        aspect-square
                        bg-parea-grey
                        bg-center
                        bg-cover
                        bg-no-repeat
                      "
                      style={{
                        backgroundImage: `url(${avatarUrl})`,
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
                        {formatUserName(user)}
                      </p>
                    </div>

                    {/* Action Button - Conditional rendering based on heading prop */}
                    {heading === 'Followers' && onRemoveFollower ? (
                      <Button
                        variant="tertiary"
                        onClick={() => handleAction(user.user_id)}
                        disabled={isActionLoading === user.user_id}
                      >
                        {isActionLoading === user.user_id ? 'Removing...' : 'Remove'}
                      </Button>
                    ) : heading === 'Following' && onUnfollow ? (
                      <Button
                        variant="tertiary"
                        onClick={() => handleAction(user.user_id)}
                        disabled={isActionLoading === user.user_id}
                      >
                        {isActionLoading === user.user_id ? 'Unfollowing...' : 'Unfollow'}
                      </Button>
                    ) : heading === 'Members' && onInvite ? (
                      <Button
                        variant="tertiary"
                        onClick={() => handleAction(user.user_id)}
                        disabled={isActionLoading === user.user_id}
                      >
                        {isActionLoading === user.user_id ? 'Inviting...' : 'Invite'}
                      </Button>
                    ) : null}
                  </div>
                );
              })
            )}
          </div>

          {/* Fade Effect Overlay */}
          {filteredUsers.length > 0 && (
            <div
              className="absolute bottom-0 left-0 right-0 h-20 pointer-events-none"
              style={{ background: 'linear-gradient(to bottom, transparent 0%, #fff 100%)' }}
            />
          )}
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