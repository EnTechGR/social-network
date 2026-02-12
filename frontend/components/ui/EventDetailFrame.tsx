'use client';

import React from 'react';
import Link from 'next/link';
import IconButton from './IconButtons';
import EventInfoBar, { type RsvpOption } from './EventInfoBar';

export interface EventDetailFrameProps {
  title: string;
  imageSrc?: string;
  eventText?: string;
  location?: string;
  goingCount?: number;
  date?: string;
  time?: string;
  rsvpValue?: RsvpOption;
  onRsvpSelect?: (option: RsvpOption) => void;
  onInviteClick?: () => void;
}

export function EventDetailFrame({
  title,
  imageSrc = '/test-event.png',
  eventText,
  location,
  goingCount = 0,
  date,
  time,
  rsvpValue,
  onRsvpSelect,
  onInviteClick,
}: EventDetailFrameProps) {
  return (
    <div className="flex flex-col items-start gap-9 self-stretch">
      {/* Inner: back button + title (same as post detail frame) */}
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

      {/* Frame: event info bar + image container (same outer styling as post) */}
      <div
        className="flex flex-col items-start self-stretch rounded border overflow-hidden"
        style={{
          height: 590,
          borderColor: 'var(--parea-border)',
          boxShadow: '8px 8px 0 0 var(--parea-black)',
        }}
      >
        {/* Holder div: event info bar instead of avatar + reactions */}
        <EventInfoBar
          location={location}
          goingCount={goingCount}
          date={date}
          time={time}
          rsvpValue={rsvpValue}
          onRsvpSelect={onRsvpSelect}
          onInviteClick={onInviteClick}
        />

        {/* Image container: same as post - flex-1, padding 0 32px 8px 32px, justify-end items-end */}
        <div className="flex flex-1 min-h-0 self-stretch px-8 pb-2 justify-end items-end gap-4">
          <img
            src={imageSrc}
            alt="Event"
            className="max-w-full max-h-full w-full h-full object-cover object-bottom rounded-[4px]"
          />
        </div>
      </div>

      {/* Details wrap: event text below image holder (no comments) */}
      {(eventText != null && eventText !== '') && (
        <div className="flex flex-col items-start self-stretch px-8">
          <div className="flex flex-col items-start self-stretch">
            <p className="text-foreground font-body text-regular font-normal leading-relaxed self-stretch">
              {eventText}
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

export default EventDetailFrame;
