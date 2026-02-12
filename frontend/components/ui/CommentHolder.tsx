'use client';

import React, { useState } from 'react';
import { CardAvatar } from './CardAvatar';
import ReactionHolder from './ReactionHolder';

export interface CommentItem {
  id: string;
  avatarSrc: string;
  avatarAlt: string;
  userName: string;
  userDate?: string;
  text: string;
  likeCount?: number;
}

export interface CommentHolderProps {
  commentValue?: string;
  onCommentChange?: (value: string) => void;
  commentPlaceholder?: string;
  comments: CommentItem[];
  className?: string;
}

export function CommentHolder({
  commentValue: controlledValue,
  onCommentChange,
  commentPlaceholder = 'Leave a comment...',
  comments,
  className = '',
}: CommentHolderProps) {
  const [internalValue, setInternalValue] = useState('');
  const isControlled = onCommentChange != null;
  const commentValue = isControlled ? (controlledValue ?? '') : internalValue;
  const handleChange = (value: string) => {
    if (isControlled) onCommentChange?.(value);
    else setInternalValue(value);
  };

  return (
    <div
      className={`flex flex-col items-start self-stretch px-8 ${className}`.trim()}
    >
      {/* Div of comment box: padding-bottom 32px */}
      <div className="flex flex-col items-start self-stretch pb-8">
        {/* Comment box: 80px height, 32px padding, border, rounded */}
        <div
          className="flex w-full min-h-[80px] flex-col justify-center items-start self-stretch rounded border p-8"
          style={{
            borderColor: 'var(--parea-border)',
            backgroundColor: 'var(--white)',
          }}
        >
          <input
            type="text"
            placeholder={commentPlaceholder}
            value={commentValue}
            onChange={(e) => handleChange(e.target.value)}
            className="w-full bg-transparent border-0 outline-none text-foreground font-body text-regular font-normal leading-relaxed placeholder:text-foreground/60"
            aria-label="Leave a comment"
          />
        </div>
      </div>

      {/* Outer container holding all comments */}
      <div className="flex flex-col items-start self-stretch rounded">
        {comments.map((comment) => (
          <div
            key={comment.id}
            className="flex flex-col items-start self-stretch pt-4 pb-8 border-b border-dashed"
            style={{ borderColor: 'rgba(18, 18, 20, 0.20)' }}
          >
            {/* Inner div of comment: flex col, gap 16px */}
            <div className="flex flex-col items-end gap-4 self-stretch">
              {/* First row: avatar */}
              <div className="flex items-start self-stretch">
                <CardAvatar
                  src={comment.avatarSrc}
                  alt={comment.avatarAlt}
                  name={comment.userName}
                  subtitle={comment.userDate}
                />
              </div>
              {/* Second row: comment text (left) + reaction (right) */}
              <div className="flex justify-between items-center gap-4 self-stretch">
                <p className="text-foreground font-body text-regular font-normal leading-relaxed flex-1 min-w-0">
                  {comment.text}
                </p>
                <ReactionHolder
                  likeCount={comment.likeCount ?? 0}
                  commentCount={0}
                  showCommentIcon={false}
                />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default CommentHolder;
