/**
 * components/ui/Button.tsx
 *
 * Reusable button component with Parea Design System styling.
 * Supports primary, secondary, and tertiary variants.
 * Tertiary buttons can be toggle buttons (with activeText/inactiveText) or regular buttons.
 */

'use client';

import React, { useState, useRef, useEffect } from 'react';

interface ButtonProps {
  children?: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'tertiary';
  size?: 'sm' | 'md' | 'lg';
  onClick?: () => void;
  disabled?: boolean;
  type?: 'button' | 'submit' | 'reset';
  className?: string;
  // For tertiary toggle button only (when activeText/inactiveText provided)
  isActive?: boolean;
  onActiveChange?: (active: boolean) => void;
  // Text for tertiary toggle button states
  activeText?: string; // Text when active (e.g., "FOLLOW")
  inactiveText?: string; // Text when inactive (e.g., "FOLLOWING")
}

export default function Button({
  children,
  variant = 'primary',
  size = 'lg',
  onClick,
  disabled = false,
  type = 'button',
  className = '',
  isActive: controlledActive,
  onActiveChange,
  activeText,
  inactiveText,
}: ButtonProps) {
  const [internalActive, setInternalActive] = useState(false);
  const [isTransitioning, setIsTransitioning] = useState(false);
  const [transitionDirection, setTransitionDirection] = useState<'toActive' | 'toInactive' | null>(null);
  const transitionTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  
  // Determine if this is a toggle button (only when activeText/inactiveText provided)
  const isToggleButton = variant === 'tertiary' && (activeText !== undefined || inactiveText !== undefined);
  
  // Use controlled or uncontrolled state (only for toggle buttons)
  const isActive = isToggleButton 
    ? (controlledActive !== undefined ? controlledActive : internalActive)
    : false;

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (transitionTimeoutRef.current) {
        clearTimeout(transitionTimeoutRef.current);
      }
    };
  }, []);

  // Base styles - different for tertiary (no padding, no border, no radius)
  const baseStyles = variant === 'tertiary'
    ? 'focus:outline-none label shrink-0 cursor-pointer p-0 border-0 rounded-none'
    : 'rounded-[3rem] border border-parea-black focus:outline-none label shrink-0 flex items-center justify-center cursor-pointer';

  const variantStyles = {
    primary: 'text-parea-black transition-colors',
    secondary: 'bg-parea-white text-parea-black hover:bg-parea-grey transition-colors',
    tertiary: 'bg-transparent text-parea-black underline', // No border, no padding
  };

  // Size styles - skip for tertiary
  const sizeStyles = variant === 'tertiary'
    ? ''
    : {
        sm: 'px-3 py-1.5 text-small',
        md: 'px-4 py-2 text-regular',
        lg: 'px-6 py-3 text-regular h-[2.9375rem]',
      }[size];

  const disabledStyles = 'opacity-50 cursor-not-allowed pointer-events-none';

  // Determine display text for tertiary toggle button
  const getDisplayText = () => {
    if (isToggleButton) {
      return isActive ? (activeText || children || '') : (inactiveText || children || '');
    }
    return children || '';
  };

  // Get transition based on direction
  const getTransition = () => {
    if (!isTransitioning || !transitionDirection) {
      return 'all 0s';
    }
    
    if (transitionDirection === 'toActive') {
      // FOLLOWING → FOLLOW: Instant
      // Mass: 1, Stiffness: 100, Damping: 15
      return 'all 0s'; // Instant = no transition
    } else {
      // FOLLOW → FOLLOWING: Spring animation
      // Mass: 1, Stiffness: 300, Damping: 20
      // Spring curve approximation: cubic-bezier(0.34, 1.56, 0.64, 1) for stiffness 300, damping 20
      return 'all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1)';
    }
  };

  // Handle click
  const handleClick = () => {
    if (disabled) return;
    
    // Handle tertiary toggle button state change
    if (isToggleButton) {
      const newActive = !isActive;
      
      // Clear any existing timeouts
      if (transitionTimeoutRef.current) {
        clearTimeout(transitionTimeoutRef.current);
      }
      
      // Set transition direction and start transition
      setTransitionDirection(newActive ? 'toActive' : 'toInactive');
      setIsTransitioning(true);
      
      // Update state
      if (controlledActive === undefined) {
        setInternalActive(newActive);
      }
      onActiveChange?.(newActive);
      
      // Stop transition after animation completes
      if (newActive) {
        // Instant transition - reset immediately
        transitionTimeoutRef.current = setTimeout(() => {
          setIsTransitioning(false);
          setTransitionDirection(null);
        }, 10);
      } else {
        // Spring animation - reset after 400ms
        transitionTimeoutRef.current = setTimeout(() => {
          setIsTransitioning(false);
          setTransitionDirection(null);
        }, 400);
      }
    }
    
    // Always call external onClick
    onClick?.();
  };

  return (
    <button
      type={type}
      onClick={handleClick}
      disabled={disabled}
      className={`
        ${baseStyles}
        ${variantStyles[variant]}
        ${sizeStyles}
        ${disabled ? disabledStyles : ''}
        ${className}
      `}
      style={{
        fontFamily: 'var(--font-ibm-plex-mono), monospace',
        fontWeight: 500,
        ...(variant === 'primary' ? { backgroundColor: '#DDFF30' } : {}),
        ...(variant === 'tertiary' ? { display: 'inline-flex', alignItems: 'center', justifyContent: 'flex-end' } : {}),
      }}
      onMouseEnter={(e) => {
        if (variant === 'primary' && !disabled) {
          e.currentTarget.style.backgroundColor = '#C8E82A';
        }
      }}
      onMouseLeave={(e) => {
        if (variant === 'primary' && !disabled) {
          e.currentTarget.style.backgroundColor = '#DDFF30';
        }
      }}
    >
      {variant === 'tertiary' && isToggleButton ? (
        // Toggle button: wrap text in span with conditional background, right-aligned
        <span
          className="inline-flex px-2 underline"
          style={{
            backgroundColor: isActive ? '#DDFF30' : 'transparent',
            justifyContent: 'flex-end', // Right align
            alignItems: 'center',
            gap: '99px',
            transition: getTransition(),
            textAlign: 'right', // Ensure text is right-aligned
          }}
        >
          {getDisplayText()}
        </span>
      ) : (
        // Regular button or non-toggle tertiary: just text
        <span>
          {getDisplayText()}
        </span>
      )}
    </button>
  );
}