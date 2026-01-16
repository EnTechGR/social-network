/**
 * components/ui/ProfileCard.tsx
 *
 * Profile card component displaying user information.
 * Used on profile pages to show avatar, name, bio, and user details.
 */

import Avatar from './Avatar';
import ToggleButton from './ToggleButton';

interface ProfileCardProps {
  /** User data to display */
  user: {
    avatarUrl?: string;
    name: string;
    username: string;
    bio?: string;
    email: string;
    birthDate: string;
    isPublic: boolean;
    followersCount: number;
    followingCount: number;
  };
  /** Whether this is the current user's own profile */
  isSelf?: boolean;
  /** Callback when public profile toggle changes */
  onTogglePublic?: (isPublic: boolean) => void;
  /** Callback when followers count is clicked */
  onFollowersClick?: () => void;
  /** Callback when following count is clicked */
  onFollowingClick?: () => void;
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
      <span className="font-medium">{label}</span>
      <span className="text-regular text-right">{children}</span>
    </div>
  );
}

/**
 * Count badge component for followers/following numbers
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
    text-small font-medium
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

export default function ProfileCard({
  user,
  isSelf = false,
  onTogglePublic,
  onFollowersClick,
  onFollowingClick,
  className = '',
}: ProfileCardProps) {
  return (
    <div
      className={`
        flex
        max-w-[1246px]
        w-full
        min-h-[309px]
        p-8
        justify-center
        items-center
        gap-24
        rounded
        border
        border-parea-border
        bg-parea-white
        shadow-[8px_8px_0_0_var(--parea-black)]
        ${className}
      `}
    >
      {/* Column 1: Avatar + Name/Bio */}
      <div className="flex items-center gap-11 flex-1 self-stretch">
        {/* Avatar */}
        <Avatar
          type={user.avatarUrl ? 'image' : 'user'}
          src={user.avatarUrl}
          alt={`${user.name}'s avatar`}
          size="xl"
          isSelf={isSelf}
        />

        {/* Name and Bio */}
        <div className="flex flex-col justify-center items-start gap-3 flex-1 self-stretch">
          <h2 className="label-lg uppercase text-parea-black">
            {user.name}
          </h2>
          {user.bio && (
            <p className="line-clamp-4 text-regular leading-relaxed text-parea-black self-stretch">
              {user.bio}
            </p>
          )}
        </div>
      </div>

      {/* Column 2: User Details */}
      <div className="flex flex-col items-center gap-3 min-w-[280px]">
        {/* Username */}
        <InfoRow label="Username">@{user.username}</InfoRow>

        {/* Email */}
        <InfoRow label="Email">{user.email}</InfoRow>

        {/* Birth Date */}
        <InfoRow label="Birth Date">{user.birthDate}</InfoRow>

        {/* Public Profile Toggle */}
        <div className="flex w-full justify-between items-center gap-4">
          <span className="font-medium">Public Profile</span>
          <ToggleButton
            isOn={user.isPublic}
            onChange={onTogglePublic}
            aria-label="Toggle public profile visibility"
          />
        </div>

        {/* Helper text for public profile */}
        <p className="text-small text-parea-black/70 self-start">
          {user.isPublic 
            ? 'Your profile can be seen by everyone.' 
            : 'Your profile is private.'}
        </p>

        {/* Followers */}
        <div className="flex w-full justify-between items-center gap-4">
          <span className="font-medium">Followers</span>
          <CountBadge count={user.followersCount} onClick={onFollowersClick} />
        </div>

        {/* Following */}
        <div className="flex w-full justify-between items-center gap-4">
          <span className="font-medium">Following</span>
          <CountBadge count={user.followingCount} onClick={onFollowingClick} />
        </div>
      </div>
    </div>
  );
}
