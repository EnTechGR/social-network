'use client';

import React, { useRef, useState } from 'react';
import Link from 'next/link';
import { CardAvatar } from './CardAvatar';
import Button from './Button';

export interface CommentItem {
  id: string;
  /** ID of the user who authored the comment (for profile link) */
  userId?: string;
  avatarSrc: string;
  avatarAlt: string;
  userName: string;
  userDate?: string;
  text: string;
  likeCount?: number;
  imageSrc?: string;
}

export interface CommentHolderProps {
  commentValue?: string;
  onCommentChange?: (value: string) => void;
  onCommentSubmit?: () => void;
  onCommentImageSelect?: (file: File) => void;
  onCommentImageRemove?: () => void;
  commentImagePreview?: string | null;
  commentImageName?: string | null;
  commentPlaceholder?: string;
  comments: CommentItem[];
  className?: string;
}

export function CommentHolder({
  commentValue: controlledValue,
  onCommentChange,
  onCommentSubmit,
  onCommentImageSelect,
  onCommentImageRemove,
  commentImagePreview,
  commentImageName,
  commentPlaceholder = 'Leave a comment...',
  comments,
  className = '',
}: CommentHolderProps) {
  const [internalValue, setInternalValue] = useState('');
  const imageInputRef = useRef<HTMLInputElement | null>(null);
  const isControlled = onCommentChange != null;
  const commentValue = isControlled ? (controlledValue ?? '') : internalValue;
  const handleChange = (value: string) => {
    if (isControlled) onCommentChange?.(value);
    else setInternalValue(value);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      if (commentValue.trim() && onCommentSubmit) {
        onCommentSubmit();
      }
    }
  };

  return (
    <div
      className={`flex flex-col items-start self-stretch px-8 ${className}`.trim()}
    >
      {/* Div of comment box: padding-bottom 32px */}
      <div className="flex flex-col items-start self-stretch pb-8">
        {/* Comment box: 80px height, 32px padding, border, rounded */}
        <div
          className="flex w-full min-h-20 gap-4 items-center self-stretch rounded border p-8"
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
            onKeyDown={handleKeyDown}
            className="flex-1 bg-transparent border-0 outline-none text-foreground font-body text-regular font-normal leading-relaxed placeholder:text-foreground/60"
            aria-label="Leave a comment"
          />
          <input
            ref={imageInputRef}
            type="file"
            accept="image/*"
            className="hidden"
            onChange={(e) => {
              const file = e.target.files?.[0];
              if (!file) return;
              onCommentImageSelect?.(file);
              e.currentTarget.value = '';
            }}
          />
          <button
            type="button"
            onClick={() => imageInputRef.current?.click()}
            className="rounded border border-parea-black px-2 py-1 text-small font-medium uppercase text-parea-black hover:opacity-90"
            aria-label="Attach image to comment"
          >
            Image
          </button>
          <Button
            variant="primary"
            size="md"
            onClick={onCommentSubmit}
            disabled={!commentValue.trim()}
          >
            POST
          </Button>
        </div>
        {commentImagePreview && (
          <div className="mt-3 flex items-center gap-2 rounded border border-parea-black/30 bg-parea-white px-3 py-2">
            <img src={commentImagePreview} alt={commentImageName || 'Comment image'} className="h-10 w-10 shrink-0 rounded object-cover" />
            <span className="max-w-55 truncate text-small text-parea-black">{commentImageName || 'Attached image'}</span>
            <button
              type="button"
              onClick={onCommentImageRemove}
              className="rounded border border-parea-black px-2 py-1 text-small uppercase text-parea-black hover:opacity-90"
            >
              Remove
            </button>
          </div>
        )}
      </div>

      {/* Outer container holding all comments */}
      <div className="flex flex-col items-start self-stretch rounded">
        {comments.map((comment) => (
          <div
            key={comment.id}
            className="flex flex-col items-start self-stretch pt-4 pb-8 border-b border-dashed"
            style={{ borderColor: 'rgba(18, 18, 20, 0.20)' }}
          >
            {/* Inner div of comment: avatar + text */}
            <div className="flex flex-col items-start gap-4 self-stretch">
              <div className="flex items-start self-stretch">
                {comment.userId ? (
                  <Link
                    href={`/profile/${comment.userId}`}
                    className="no-underline text-inherit hover:opacity-90 transition-opacity"
                  >
                    <CardAvatar
                      src={comment.avatarSrc}
                      alt={comment.avatarAlt}
                      name={comment.userName}
                      subtitle={comment.userDate}
                    />
                  </Link>
                ) : (
                  <CardAvatar
                    src={comment.avatarSrc}
                    alt={comment.avatarAlt}
                    name={comment.userName}
                    subtitle={comment.userDate}
                  />
                )}
              </div>
              <div className="flex self-stretch">
                <div className="flex-1 min-w-0">
                  <p className="text-foreground font-body text-regular font-normal leading-relaxed">
                    {comment.text}
                  </p>
                  {comment.imageSrc && (
                    <img
                      src={comment.imageSrc}
                      alt="Comment attachment"
                      className="mt-3 max-h-56 w-auto max-w-full rounded border border-parea-black/20 object-contain"
                    />
                  )}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default CommentHolder;
