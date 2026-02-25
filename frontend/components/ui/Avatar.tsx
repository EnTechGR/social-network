/**
 * components/ui/Avatar.tsx
 *
 * Reusable avatar component with Parea Design System styling.
 * Supports self/not-self states and three avatar types.
 */

import React from 'react';
import Image from 'next/image';

interface AvatarProps {
  /** Whether this avatar represents the current user (self) */
  isSelf?: boolean;
  /** Type of avatar: user (single person), group, or image (custom uploaded) */
  type?: 'user' | 'group' | 'image';
  /** Image source URL for 'image' type avatars */
  src?: string;
  /** Alt text for accessibility */
  alt?: string;
  /** Size variant */
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'profile';
  /** Additional CSS classes */
  className?: string;
}

export default function Avatar({
  isSelf = false,
  type = 'user',
  src,
  alt = 'Avatar',
  size = 'lg',
  className = '',
}: AvatarProps) {
  // Base styles: black 1px stroke
  const baseStyles = 'rounded-[4px] border border-parea-black bg-parea-yellow flex items-center justify-center overflow-hidden';

  // Shadow: isSelf → yellow (4px x, 4px y); not self → black (4px x, 4px y)
  const shadowStyles = isSelf
    ? 'shadow-[4px_4px_0_0_var(--parea-yellow)]'
    : 'shadow-[4px_4px_0_0_#000]';

  // Size variants
  const sizeStyles = {
    sm: 'w-10 h-10',       // 40px
    md: 'w-20 h-20',       // 80px
    lg: 'w-40 h-40',       // 160px
    xl: 'w-60 h-60',       // 240px - for profile cards
    profile: 'w-32 h-32',   // 128px - signup / profile preview
  };

  // When type is 'image' and no src = placeholder (e.g. signup): yellow + avatar.png only, no default image (avoids dots from user-avatar-default.png)
  const isPlaceholderNoImage = type === 'image' && !src;

  const renderContent = () => {
    // type='image' with no src = placeholder only (yellow + avatar.png overlay); never use default user/group image (no dots)
    if (type === 'image' && !src) return null;

    // If src is provided, use it (unoptimized for data URLs e.g. file preview)
    if (src) {
      const isDataUrl = src.startsWith('data:');
      return (
        <Image
          src={src}
          alt={alt}
          fill
          className="object-cover"
          unoptimized={isDataUrl}
        />
      );
    }

    // Groups should not render a default avatar image when no source is provided.
    if (type === 'group') {
      return null;
    }

    // Default image for user avatars when no source exists.
    const defaultSrc = '/user-avatar-default.png';

    return (
      <Image
        src={defaultSrc}
        alt={alt}
        fill
        className="object-cover"
      />
    );
  };

  return (
    <div
      className={`
        ${baseStyles}
        ${shadowStyles}
        ${sizeStyles[size]}
        ${className}
        relative
      `}
      role="img"
      aria-label={alt}
    >
      {/* Avatar image (user photo or default); placeholder (type=image, no src) shows only yellow + frame below */}
      <div className="absolute inset-0">{renderContent()}</div>
      {/* Frame overlay only when no custom image (placeholder state) */}
      {isPlaceholderNoImage && (
        <Image
          src="/avatar.png"
          alt=""
          fill
          className="object-contain pointer-events-none"
          aria-hidden
        />
      )}
    </div>
  );
}