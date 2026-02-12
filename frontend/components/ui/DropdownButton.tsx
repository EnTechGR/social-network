'use client';

import React, { useState, useRef, useEffect } from 'react';

export interface DropdownButtonProps {
  /** List of options (displayed as-is, typically uppercase) */
  options: string[];
  /** Controlled value: the currently selected option */
  value?: string;
  /** Uncontrolled default (used when value is not provided) */
  defaultValue?: string;
  /** Called when user selects an option */
  onValueChange?: (value: string) => void;
  /** Accessibility label for the trigger button */
  'aria-label'?: string;
  className?: string;
}

const optionBaseClass =
  'block w-full px-5 py-2 text-left text-base font-medium uppercase tracking-[-0.01em] leading-relaxed hover:bg-parea-grey transition-colors cursor-pointer';
const optionFontStyle = { fontFamily: 'var(--font-ibm-plex-mono), monospace' as const };

export function DropdownButton({
  options,
  value: controlledValue,
  defaultValue,
  onValueChange,
  'aria-label': ariaLabel,
  className = '',
}: DropdownButtonProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [internalValue, setInternalValue] = useState(defaultValue ?? options[0] ?? '');
  const dropdownRef = useRef<HTMLDivElement>(null);

  const isControlled = controlledValue !== undefined;
  const displayValue = isControlled ? controlledValue : internalValue;

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSelect = (option: string) => {
    if (!isControlled) setInternalValue(option);
    onValueChange?.(option);
    setIsOpen(false);
  };

  if (options.length === 0) return null;

  return (
    <div ref={dropdownRef} className={`relative ${className}`.trim()}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center justify-between w-40 gap-2 px-5 py-2 bg-white border border-[#222] cursor-pointer shadow-[0.25rem_0.25rem_0_0_#000]"
        style={optionFontStyle}
        aria-expanded={isOpen}
        aria-haspopup="listbox"
        aria-label={ariaLabel ?? 'Select option'}
      >
        <span className="text-base font-medium uppercase tracking-[-0.01em] leading-relaxed">
            {displayValue}
          </span>
          <svg
            width="12"
            height="7"
            viewBox="0 0 12 7"
            fill="none"
            className={`shrink-0 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`}
            aria-hidden
          >
          <path
            d="M1 1L6 6L11 1"
            stroke="#000000"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </button>

      {isOpen && (
        <div
          className="absolute top-full left-0 w-40 bg-white border border-[#222] border-t-0 z-10 shadow-[0.25rem_0.25rem_0_0_#000]"
          role="listbox"
        >
          {options.map((option) => (
            <button
              key={option}
              type="button"
              role="option"
              onClick={() => handleSelect(option)}
              className={`${optionBaseClass} ${displayValue === option ? 'bg-parea-grey' : ''}`}
              style={optionFontStyle}
            >
              {option}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

export default DropdownButton;
