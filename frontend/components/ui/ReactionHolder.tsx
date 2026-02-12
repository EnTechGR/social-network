'use client';

import React from 'react';
import { Heart, MessageCircle } from 'lucide-react';

export interface ReactionHolderProps {
  likeCount?: number;
  commentCount?: number;
  showCommentIcon?: boolean;
  onLikeClick?: () => void;
  onCommentClick?: () => void;
  className?: string;
}

export function ReactionHolder({
  likeCount = 0,
  commentCount = 0,
  showCommentIcon = true,
  onLikeClick,
  onCommentClick,
  className = '',
}: ReactionHolderProps) {
  return (
    <div
      className={`flex items-center gap-4 text-foreground ${className}`.trim()}
      role="group"
      aria-label="Reactions"
    >
      <span className="flex items-center gap-2 text-sm font-body">
        {onLikeClick ? (
          <button
            type="button"
            onClick={onLikeClick}
            className="flex items-center gap-2 p-0 border-0 bg-transparent cursor-pointer text-foreground hover:opacity-80"
            aria-label={`${likeCount} likes`}
          >
            <Heart className="w-6 h-6 shrink-0" strokeWidth={1.5} />
            {likeCount}
          </button>
        ) : (
          <>
            <Heart className="w-6 h-6 shrink-0" strokeWidth={1.5} />
            {likeCount}
          </>
        )}
      </span>
      {showCommentIcon && (
        <span className="flex items-center gap-2 text-sm font-body">
          {onCommentClick ? (
            <button
              type="button"
              onClick={onCommentClick}
              className="flex items-center gap-2 p-0 border-0 bg-transparent cursor-pointer text-foreground hover:opacity-80"
              aria-label={`${commentCount} comments`}
            >
              <MessageCircle className="w-6 h-6 shrink-0" strokeWidth={1.5} />
              {commentCount}
            </button>
          ) : (
            <>
              <MessageCircle className="w-6 h-6 shrink-0" strokeWidth={1.5} />
              {commentCount}
            </>
          )}
        </span>
      )}
    </div>
  );
}

export default ReactionHolder;
