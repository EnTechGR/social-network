/**
 * components/ui/GroupWrap.tsx
 *
 * Group card component displaying group information.
 * Used on group pages to show image, name, description, and group details.
 * Has two states: non-member (Join button) and member (Invite + Leave buttons)
 */

'use client';

import { useEffect, useState } from 'react';
import Button from './Button';
import Avatar from './Avatar';

interface GroupWrapProps {
  /** Group data to display */
  group: {
    imageUrl?: string;
    name: string;
    description?: string;
    admin: string;
    createdDate: string;
    membersCount: number;
  };
  /** Whether the current user is a member of the group */
  isMember: boolean;
  /** Whether a join request has been sent (for pending state) */
  isPending?: boolean;
  /** Callback when Join button is clicked */
  onJoin?: () => void;
  /** Callback when Invite button is clicked */
  onInvite?: () => void;
  /** Callback when Leave Group button is clicked */
  onLeaveGroup?: () => void;
  /** Callback when Members count is clicked */
  onMembersClick?: () => void;
  /** Additional CSS classes */
  className?: string;
}

/**
 * Info row component for displaying label-value pairs
 */
function InfoRow({ 
  label, 
  children 
}: { 
  label: string; 
  children: React.ReactNode;
}) {
  return (
    <div className="flex w-full justify-between items-center gap-4">
      <span className="font-weight-medium">{label}</span>
      <span className="text-regular text-right">{children}</span>
    </div>
  );
}

/**
 * Count badge component for members count
 */
function CountBadge({ 
  count, 
  onClick 
}: { 
  count: number; 
  onClick?: () => void;
}) {
  const baseStyles = `
    px-3 py-1 
    rounded-full 
    border border-parea-black 
    bg-parea-yellow 
    text-small font-mono font-medium
    tabular-nums
  `;
  
  if (onClick) {
    return (
      <button
        type="button"
        onClick={onClick}
        className={`${baseStyles} cursor-pointer hover:opacity-80 transition-opacity`}
      >
        {count}
      </button>
    );
  }
  
  return <span className={baseStyles}>{count}</span>;
}

export default function GroupWrap({
  group,
  isMember = false,
  isPending: initialPending = false,
  onJoin,
  onInvite,
  onLeaveGroup,
  onMembersClick,
  className = '',
}: GroupWrapProps) {
  // Local state for pending join request
  const [isPending, setIsPending] = useState(initialPending);

  useEffect(() => {
    setIsPending(initialPending);
  }, [initialPending]);

  const handleJoinClick = () => {
    setIsPending(true);
    if (onJoin) {
      onJoin();
    }
  };

  return (
    <div
      className={`
        flex
        flex-col lg:flex-row
        max-w-[1246px]
        w-full
        min-h-[309px]
        p-6 lg:p-8
        justify-center
        items-start
        gap-6 md:gap-16 lg:gap-24
        rounded
        border
        border-parea-border
        bg-parea-white
        shadow-[8px_8px_0_0_var(--parea-black)]
        ${className}
      `}
    >
      {/* Column 1: Image + Name/Description */}
      <div className="flex flex-col lg:flex-row items-center lg:items-start gap-6 lg:gap-11 flex-1 self-stretch">
        {/* Group Image */}
        <Avatar
            type={group.imageUrl ? 'image' : 'group'}
            src={group.imageUrl}
            alt={`${group.name}`}
            size="xl"
        />

        {/* Name and Description */}
        <div className="flex flex-col justify-center items-start gap-3 flex-1 self-stretch">
          <h2 className="label-lg uppercase text-parea-black">
            {group.name}
          </h2>
          {group.description && (
            <p className="line-clamp-4 text-regular leading-relaxed text-parea-black self-stretch">
              {group.description}
            </p>
          )}
        </div>
      </div>

      {/* Column 2: Buttons + Group Info */}
      <div className="flex flex-col items-start lg:items-end justify-between min-w-[280px] self-stretch">
        {/* Action Buttons */}
        <div className="flex gap-2 mb-4">
          {isMember ? (
            <>
              <Button variant="primary" size="sm" onClick={onInvite}>
                Invite
              </Button>
              <Button variant="secondary" size="sm" onClick={onLeaveGroup}>
                Leave Group
              </Button>
            </>
          ) : (
            <div className="flex items-center gap-2 lg:flex-col lg:items-end">
              <Button 
                variant="primary" 
                size="sm" 
                onClick={handleJoinClick}
                disabled={isPending}
                className={isPending ? 'opacity-40' : ''}
              >
                Join
              </Button>
              {isPending && (
                <span className="text-small text-parea-black/60 whitespace-nowrap lg:mt-1">
                  Sent request.
                </span>
              )}
            </div>
          )}
        </div>

        {/* Admin */}
        <div className="flex flex-col gap-3 w-full">
        <InfoRow label="Admin">{group.admin}</InfoRow>

        {/* Creating Date */}
        <InfoRow label="Creating Date">{group.createdDate}</InfoRow>

        {/* Members */}
        <div className="flex w-full justify-between items-center gap-4">
          <span className="font-size-medium">Members</span>
          <CountBadge count={group.membersCount} onClick={onMembersClick} />
        </div>
        </div>
      </div>
    </div>
  );
}
