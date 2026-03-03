import React from 'react';
import Image from 'next/image';

interface CardAvatarProps {
  src?: string;
  alt: string;
  name: string;
  subtitle?: string;
}

const DEFAULT_AVATAR_SRC = '/user-avatar-default.png';

export const CardAvatar: React.FC<CardAvatarProps> = ({
  src,
  alt,
  name,
  subtitle,
}) => {
  return (
    <div className="flex
      h-12
      items-end
      gap-4
      self-stretch">
      <div className="relative w-12 h-12 rounded-full shrink-0 aspect-square overflow-hidden bg-parea-grey">
        <Image
          src={src || DEFAULT_AVATAR_SRC}
          alt={alt}
          fill
          className="object-cover object-center"
        />
      </div>
      <div className="flex
        flex-col
        items-start
        flex-1
        self-stretch">
        <p
          className="self-stretch
            text-black
            text-base
            font-medium
            uppercase"
          style={{
            fontFamily: '"IBM Plex Mono"',
            lineHeight: '150%',
            letterSpacing: '-0.16px',
          }}>
          {name}
        </p>
        {subtitle && (
          <div className="flex
            items-center
            gap-2
            self-stretch">
            <p
              className="text-black
                text-sm
                font-normal"
              style={{
                fontFamily: 'Inter',
                lineHeight: '150%',
              }}>
              {subtitle}
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export default CardAvatar;
