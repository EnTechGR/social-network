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
  size?: 'sm' | 'md' | 'lg';
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
  };

  // Self vs Not-self shadow styles
  // Self: yellow shadow | Not-self: black shadow
  const shadowStyles = isSelf
    ? 'shadow-[4px_4px_0_0_var(--parea-yellow)]'
    : 'shadow-[4px_4px_0_0_#000]';

  // Render the appropriate avatar content based on type
  const renderContent = () => {
    if (type === 'image' && src) {
      return (
        <Image
          src={src}
          alt={alt}
          fill
          className="object-cover"
        />
      );
    }

    if (type === 'group') {
      // Default group avatar icon (you'd replace with your actual SVG)
      return (
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
          className="w-1/2 h-1/2 text-parea-black"
        >
          <path d="M18 18.86h-.76c-.8 0-1.2 0-1.53.12a2 2 0 0 0-1.08 1.04c-.14.32-.17.71-.2 1.48M18 18.86c1.13-.47 2-1.4 2-2.86 0-2-1.79-3-4-3M14 6.12c.24-.08.5-.12.77-.12 1.8 0 3.23 1.57 3.23 3.5s-1.44 3.5-3.23 3.5c-.27 0-.53-.04-.77-.12M6 18.86h.76c.8 0 1.2 0 1.53.12.46.17.82.52 1.08 1.04.14.32.17.71.2 1.48M6 18.86c-1.13-.47-2-1.4-2-2.86 0-2 1.79-3 4-3s4 1 4 3c0 1.46-.87 2.39-2 2.86m0 0h-.76c-.8 0-1.2 0-1.53.12a2 2 0 0 0-1.08 1.04c-.14.32-.17.71-.2 1.48M11.5 6.5c0 1.93-1.44 3.5-3.23 3.5S5.04 8.43 5.04 6.5 6.48 3 8.27 3s3.23 1.57 3.23 3.5Z" />
        </svg>
      );
    }

    // Default user avatar icon
    return (
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        className="w-1/2 h-1/2 text-parea-black"
      >
        <circle cx="12" cy="8" r="4" />
        <path d="M4 20c0-2.5 3.5-4 8-4s8 1.5 8 4" />
      </svg>
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