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
  size?: 'sm' | 'md' | 'lg' | 'xl';
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
  // Base styles shared by all avatars
  const baseStyles = 'rounded-[4px] border border-parea-black bg-parea-yellow flex items-center justify-center overflow-hidden';

  // Size variants
  const sizeStyles = {
    sm: 'w-10 h-10',     // 40px
    md: 'w-20 h-20',     // 80px  
    lg: 'w-40 h-40',     // 160px
    xl: 'w-60 h-60',     // 240px - for profile cards
  };

  // Self vs Not-self shadow styles
  // Self: yellow shadow | Not-self: black shadow
  const shadowStyles = isSelf
    ? 'shadow-[4px_4px_0_0_var(--parea-yellow)]'
    : 'shadow-[4px_4px_0_0_#000]';

  // Render the appropriate avatar content based on type
  const renderContent = () => {
    // If src is provided, use it
    if (src) {
      return (
        <Image
          src={src}
          alt={alt}
          fill
          className="object-cover"
        />
      );
    }
  
    // Use default image based on type
    const defaultSrc = type === 'group' 
      ? '/group-avatar-default.png' 
      : '/user-avatar-default.png';
  
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
        ${sizeStyles[size]}
        ${shadowStyles}
        ${className}
        relative
      `}
      role="img"
      aria-label={alt}
    >
      {renderContent()}
    </div>
  );
}