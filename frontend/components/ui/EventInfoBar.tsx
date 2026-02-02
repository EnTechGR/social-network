'use client';

import React, { useState } from 'react';
import { MapPin, Calendar, Clock } from 'lucide-react';
import Button from './Button';
import DropdownButton from './DropdownButton';

export type RsvpOption = 'RSVP' | 'GOING' | "CAN'T GO";

export interface EventInfoBarProps {
  location?: string;
  goingCount?: number;
  date?: string;
  time?: string;
  rsvpValue?: RsvpOption;
  onRsvpSelect?: (option: RsvpOption) => void;
  onInviteClick?: () => void;
  className?: string;
}

const labelClass = 'label text-foreground';

const RSVP_OPTIONS: RsvpOption[] = ['RSVP', 'GOING', "CAN'T GO"];

export function EventInfoBar({
  location,
  goingCount = 0,
  date,
  time,
  rsvpValue,
  onRsvpSelect,
  onInviteClick,
  className = '',
}: EventInfoBarProps) {
  const [selectedRsvp, setSelectedRsvp] = useState<RsvpOption>(rsvpValue ?? 'RSVP');
  const isControlled = rsvpValue !== undefined;
  const displayRsvp = isControlled ? rsvpValue : selectedRsvp;

  const handleRsvpChange = (option: string) => {
    if (!isControlled) setSelectedRsvp(option as RsvpOption);
    onRsvpSelect?.(option as RsvpOption);
  };

  return (
    <div
      className={`flex min-h-24 shrink-0 flex-col gap-4 self-stretch px-8 py-6 bg-parea-white md:flex-row md:items-center md:justify-between ${className}`.trim()}
    >
      {/* Div holding event info: wraps with consistent gap; row layout only at md+ */}
      <div className="flex flex-wrap items-end gap-x-6 gap-y-3 md:gap-x-12 md:gap-y-0">
        {location != null && (
          <div className="flex items-start gap-1">
            <MapPin className="mt-0.5 h-4 w-4 shrink-0 text-foreground" strokeWidth={1.5} />
            <span className={labelClass}>{location}</span>
          </div>
        )}
        <div className="flex items-start gap-1">
          <span
            className="text-foreground font-bold leading-relaxed uppercase tracking-[-0.15px]"
            style={{
              fontFamily: 'var(--font-ibm-plex-mono), monospace',
              fontSize: '15px',
              WebkitTextStroke: '0.3px #000',
            }}
          >
            {goingCount}
          </span>
          <span className={labelClass}>GOING</span>
        </div>
        {date != null && (
          <div className="flex items-center gap-1">
            <Calendar className="h-4 w-4 shrink-0 text-foreground" strokeWidth={1.5} />
            <span className={labelClass}>{date}</span>
          </div>
        )}
        {time != null && (
          <div className="flex items-center gap-1">
            <Clock className="h-4 w-4 shrink-0 text-foreground" strokeWidth={1.5} />
            <span className={labelClass}>{time}</span>
          </div>
        )}
      </div>

      {/* Div holding two buttons: under event info below md, right on md+ */}
      <div className="flex shrink-0 items-center justify-start gap-4 md:justify-end">
        <DropdownButton
          options={RSVP_OPTIONS}
          value={displayRsvp}
          onValueChange={handleRsvpChange}
          aria-label="RSVP options"
        />
        <Button variant="primary" size="sm" onClick={onInviteClick}>
          INVITE
        </Button>
      </div>
    </div>
  );
}

export default EventInfoBar;
