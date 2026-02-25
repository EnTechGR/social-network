import React from 'react';
import Link from 'next/link';
import { CardImage } from './CardImage';
import { CardAvatar } from './CardAvatar';
import ReactionHolder from './ReactionHolder';

interface CardProps {
  imageType: 'post' | 'event';
  imageSrc?: string;
  hideImage?: boolean;
  avatarSrc: string;
  avatarAlt: string;
  userName: string;
  userDate?: string;
  /** Optional link to the author's profile (e.g. /profile/[id]) */
  userHref?: string;
  /** When provided, card title (e.g. post title) */
  title?: string;
  /** When provided, card body (e.g. post content) */
  content?: string;
  /** When provided, entire card links to this href (e.g. /post/[id]) */
  href?: string;
  /** Set for first card in feed to fix LCP (loading="eager") */
  imagePriority?: boolean;
  /** Like count for the post */
  likeCount?: number;
  /** Comment count for the post */
  commentCount?: number;
}

export const Card: React.FC<CardProps> = ({
  imageType,
  imageSrc,
  hideImage = false,
  avatarSrc,
  avatarAlt,
  userName,
  userDate,
  userHref,
  title,
  content,
  href,
  imagePriority,
  likeCount = 0,
  commentCount = 0,
}) => {
  const inner = (
    <div className="flex
      w-full
      max-w-[917px]
      min-h-[206px]
      pb-8
      items-center
      gap-4
      border-b
      border-dashed
      border-black/20">
      {!hideImage && (
        <CardImage
          type={imageType}
          src={imageSrc}
          alt="Card image"
          priority={imagePriority}
        />
      )}
        <div className={`flex
        ${hideImage ? 'px-0' : 'px-8'}
        flex-col
        justify-between
        items-start
        flex-1
        self-stretch`}>
        <div className="flex
          flex-col
          items-start
          gap-4
          self-stretch">
          <h3 className="text-lg font-semibold text-black">
            {title ?? 'Discover Amazing Content'}
          </h3>
          <p className="text-sm text-gray-700 line-clamp-3">
            {content ?? 'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.'}
          </p>
        </div>
        <div className="flex justify-between items-center w-full self-stretch mt-4">
          {userHref && !href ? (
            <Link
              href={userHref}
              className="no-underline text-inherit hover:opacity-90 transition-opacity"
              onClick={(e) => {
                // Prevent the parent card link (href) from firing when clicking the avatar
                e.stopPropagation();
              }}
            >
              <CardAvatar
                src={avatarSrc}
                alt={avatarAlt}
                name={userName}
                subtitle={userDate}
              />
            </Link>
          ) : (
            <CardAvatar
              src={avatarSrc}
              alt={avatarAlt}
              name={userName}
              subtitle={userDate}
            />
          )}
          <ReactionHolder
            commentCount={commentCount}
          />
        </div>
      </div>
    </div>
  );

  if (href) {
    return (
      <Link
        href={href}
        className="block w-full self-stretch no-underline text-inherit hover:opacity-95 transition-opacity"
      >
        {inner}
      </Link>
    );
  }

  return inner;
};

export default Card;
