import React from 'react';
import { CardImage } from './CardImage';
import { CardAvatar } from './CardAvatar';

interface CardProps {
  imageType: 'post' | 'event';
  imageSrc?: string;
  avatarSrc: string;
  avatarAlt: string;
  userName: string;
  userDate?: string;
}

export const Card: React.FC<CardProps> = ({
  imageType,
  imageSrc,
  avatarSrc,
  avatarAlt,
  userName,
  userDate,
}) => {
  return (
    <div className="flex
      w-[917px]
      h-[306px]
      pb-8
      items-center
      gap-4
      border-b
      border-dashed
      border-black/20">
      <CardImage
        type={imageType}
        src={imageSrc}
        alt="Card image"
      />
      <div className="flex
        px-8
        flex-col
        justify-between
        items-start
        flex-1
        self-stretch">
        <div className="flex
          flex-col
          items-start
          gap-4
          self-stretch">
          <h3 className="text-lg font-semibold text-black">
            Discover Amazing Content
          </h3>
          <p className="text-sm text-gray-700 line-clamp-3">
            Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
          </p>
        </div>
        <CardAvatar
          src={avatarSrc}
          alt={avatarAlt}
          name={userName}
          subtitle={userDate}
        />
      </div>
    </div>
  );
};

export default Card;
