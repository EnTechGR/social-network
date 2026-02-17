'use client';

import React from 'react';
import Link from 'next/link';
import IconButton from './IconButtons';
import { CardAvatar } from './CardAvatar';
import ReactionHolder from './ReactionHolder';
import CommentHolder, { type CommentItem } from './CommentHolder';

export interface PostDetailFrameProps {
  title: string;
  avatarSrc: string;
  avatarAlt: string;
  userName: string;
  userDate?: string;
  likeCount?: number;
  commentCount?: number;
  imageSrc?: string;
  postText?: string;
  comments?: CommentItem[];
  commentValue?: string;
  onCommentChange?: (value: string) => void;
  onCommentSubmit?: () => void;
}

export function PostDetailFrame({
  title,
  avatarSrc,
  avatarAlt,
  userName,
  userDate,
  likeCount = 250,
  commentCount = 4,
  imageSrc,
  postText,
  comments = [],
  commentValue,
  onCommentChange,
  onCommentSubmit,
}: PostDetailFrameProps) {
  return (
    <div className="flex flex-col items-start gap-9 self-stretch">
      {/* Inner: back button + title */}
      <div className="flex flex-col items-start gap-6 self-stretch">
        <div className="flex justify-start items-center gap-2 self-stretch">
          <Link href="/feed" aria-label="Back to feed">
            <IconButton
              variant="arrow-left"
              aria-label="Back to feed"
              text="BACK"
            />
          </Link>
        </div>
        <h2 className="text-foreground font-body font-bold self-stretch text-h2 leading-tight">
          {title}
        </h2>
      </div>

      {/* Frame: avatar row + image container */}
      {imageSrc ? (
        <div
          className="flex flex-col items-start self-stretch rounded border overflow-hidden"
          style={{
            height: 590,
            borderColor: 'var(--parea-border)',
            boxShadow: '8px 8px 0 0 var(--parea-black)',
          }}
        >
          {/* Top row: avatar (left) + like/comment (right); on mobile reactions under avatar */}
          <div className="flex flex-col gap-4 w-full px-8 pt-6 pb-4 sm:flex-row sm:items-center sm:justify-between">
            <CardAvatar
              src={avatarSrc}
              alt={avatarAlt}
              name={userName}
              subtitle={userDate}
            />
            <ReactionHolder likeCount={likeCount} commentCount={commentCount} />
          </div>

          {/* Image container: flex-1, padding 0 32px 8px 32px, justify-end items-end */}
          <div className="flex flex-1 min-h-0 self-stretch px-8 pb-2 justify-end items-end gap-4">
            <img
              src={imageSrc}
              alt="Post"
              className="max-w-full max-h-full w-full h-full object-cover object-bottom rounded-[4px]"
            />
          </div>
        </div>
      ) : (
        <div className="flex flex-col items-start self-stretch rounded border overflow-hidden p-8">
          <div className="flex flex-col gap-4 w-full sm:flex-row sm:items-center sm:justify-between">
            <CardAvatar
              src={avatarSrc}
              alt={avatarAlt}
              name={userName}
              subtitle={userDate}
            />
            <ReactionHolder likeCount={likeCount} commentCount={commentCount} />
          </div>
        </div>
      )}

      {/* Details wrap: post text below image holder */}
      {(postText != null && postText !== '') && (
        <div className="flex flex-col items-start self-stretch px-8">
          <div className="flex flex-col items-start self-stretch">
            <p className="text-foreground font-body text-regular font-normal leading-relaxed self-stretch">
              {postText}
            </p>
          </div>
        </div>
      )}

      {/* Comment holder: comment box + list of comments */}
      <CommentHolder
        commentValue={commentValue}
        onCommentChange={onCommentChange}
        onCommentSubmit={onCommentSubmit}
        comments={comments}
      />
    </div>
  );
}

export default PostDetailFrame;
