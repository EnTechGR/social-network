'use client';

import { useState } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface DatePickerProps {
  selectedDate?: Date;
  onSelectDate?: (date: Date) => void;
}

const DAYS_OF_WEEK = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

const MONTHS = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December'
];

function getDaysInMonth(year: number, month: number): number {
  return new Date(year, month + 1, 0).getDate();
}

function getFirstDayOfMonth(year: number, month: number): number {
  const day = new Date(year, month, 1).getDay();
  return day === 0 ? 6 : day - 1; // Convert to Monday-first
}

function isSameDay(date1: Date, date2: Date): boolean {
  return (
    date1.getDate() === date2.getDate() &&
    date1.getMonth() === date2.getMonth() &&
    date1.getFullYear() === date2.getFullYear()
  );
}

export default function DatePicker({ selectedDate, onSelectDate }: DatePickerProps) {
  const today = new Date();
  const [viewDate, setViewDate] = useState(selectedDate || today);

  const year = viewDate.getFullYear();
  const month = viewDate.getMonth();
  const daysInMonth = getDaysInMonth(year, month);
  const firstDay = getFirstDayOfMonth(year, month);

  // Previous month days to fill first row
  const prevMonthDays = getDaysInMonth(year, month - 1);
  const prevMonthFill = Array.from({ length: firstDay }, (_, i) => prevMonthDays - firstDay + 1 + i);

  // Current month days
  const currentMonthDays = Array.from({ length: daysInMonth }, (_, i) => i + 1);

  // Next month days to fill last row
  const totalCells = Math.ceil((firstDay + daysInMonth) / 7) * 7;
  const nextMonthFill = Array.from({ length: totalCells - firstDay - daysInMonth }, (_, i) => i + 1);

  const handleSelectDay = (day: number) => {
    const newDate = new Date(year, month, day);
    onSelectDate?.(newDate);
  };

  const isSelected = (day: number): boolean => {
    if (!selectedDate) return false;
    return isSameDay(selectedDate, new Date(year, month, day));
  };

  const isToday = (day: number): boolean => {
    return isSameDay(today, new Date(year, month, day));
  };

  return (
    <div className="bg-parea-white border border-parea-black p-4 shadow-[8px_8px_0_0_#000] w-53">
      {/* Month Navigation */}
      <div className="flex items-center justify-between mb-2">
        <button
          type="button"
          onClick={() => setViewDate(new Date(year, month - 1, 1))}
          className="p-1 hover:bg-parea-black/10 rounded transition-colors cursor-pointer"
        >
          <ChevronLeft size={20} />
        </button>
        <span className="text-sm font-normal">
          {MONTHS[month]} {year}
        </span>
        <button
          type="button"
          onClick={() => setViewDate(new Date(year, month + 1, 1))}
          className="p-1 hover:bg-parea-black/10 rounded transition-colors cursor-pointer"
        >
          <ChevronRight size={20} />
        </button>
      </div>

      {/* Days of Week Header */}
      <div className="grid grid-cols-7 mb-1">
        {DAYS_OF_WEEK.map((day) => (
          <div key={day} className="text-center text-[8px] text-parea-black py-1">
            {day}
          </div>
        ))}
      </div>

      {/* Calendar Grid */}
      <div className="grid grid-cols-7">
        {/* Previous month days (faded) */}
        {prevMonthFill.map((day) => (
          <div
            key={`prev-${day}`}
            className="h-6.5 flex items-center justify-center text-[13px] text-parea-black/30"
          >
            {day}
          </div>
        ))}

        {/* Current month days */}
        {currentMonthDays.map((day) => (
          <button
            key={day}
            type="button"
            onClick={() => handleSelectDay(day)}
            className={`h-6.5 w-full flex items-center justify-center text-[13px] transition-colors cursor-pointer
              ${isSelected(day) ? 'bg-parea-yellow text-parea-black font-medium' : 'hover:bg-parea-black/10'}
              ${isToday(day) ? 'border border-parea-black/30' : ''}
            `}
          >
            {day}
          </button>
        ))}

        {/* Next month days (faded) */}
        {nextMonthFill.map((day) => (
          <div
            key={`next-${day}`}
            className="h-6.5 flex items-center justify-center text-[13px] text-parea-black/30"
          >
            {day}
          </div>
        ))}
      </div>
    </div>
  );
}
