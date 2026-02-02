'use client';

import { useRef, useEffect } from 'react';

interface TimePickerProps {
  selectedTime?: string;
  onSelectTime?: (time: string) => void;
}

// Generate time slots from 12:00 AM to 11:00 PM
function generateTimeSlots(): string[] {
  const slots: string[] = [];
  for (let hour = 0; hour < 24; hour++) {
    const h = hour % 12 || 12;
    const period = hour < 12 ? 'AM' : 'PM';
    slots.push(`${h.toString().padStart(2, '0')}:00 ${period}`);
  }
  return slots;
}

const TIME_SLOTS = generateTimeSlots();

export default function TimePicker({ selectedTime, onSelectTime }: TimePickerProps) {
  const listRef = useRef<HTMLDivElement>(null);
  const selectedRef = useRef<HTMLButtonElement>(null);

  // Scroll to selected time on mount
  useEffect(() => {
    if (selectedRef.current && listRef.current) {
      selectedRef.current.scrollIntoView({ block: 'nearest' });
    }
  }, []);

  return (
    <div className="bg-parea-white border border-parea-black shadow-[8px_8px_0_0_#000] w-35">
      <div
        ref={listRef}
        className="flex flex-col max-h-58 overflow-y-auto"
      >
        {TIME_SLOTS.map((time) => {
          const isSelected = time === selectedTime;
          return (
            <button
              key={time}
              ref={isSelected ? selectedRef : null}
              type="button"
              onClick={() => onSelectTime?.(time)}
              className={`px-5 py-2 text-left text-base transition-colors cursor-pointer
                ${isSelected ? 'bg-parea-yellow' : 'hover:bg-parea-black/10'}
              `}
            >
              {time}
            </button>
          );
        })}
      </div>
    </div>
  );
}
