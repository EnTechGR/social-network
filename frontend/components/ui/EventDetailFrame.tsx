'use client';

import React from 'react';
import Link from 'next/link';
import IconButton from './IconButtons';
import EventInfoBar, { type RsvpOption } from './EventInfoBar';

export interface EventDetailFrameProps {
  title: string;
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

      {/* Frame: event info bar */}
      <div
        className="flex flex-col items-start self-stretch rounded border overflow-hidden"
        style={{
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

      </div>

      {/* Details wrap: event text below info bar */}
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
