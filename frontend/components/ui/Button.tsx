/**
 * components/ui/Button.tsx
 *
 * Reusable button component with Parea Design System styling.
 */

import React from 'react';

interface ButtonProps {
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'tertiary';
  size?: 'sm' | 'md' | 'lg';
  onClick?: () => void;
  disabled?: boolean;
  type?: 'button' | 'submit' | 'reset';
  className?: string;
}

export default function Button({
  children,
  variant = 'primary',
  size = 'lg',
  onClick,
  disabled = false,
  type = 'button',
  className = '',
}: ButtonProps) {
  const baseStyles = 'rounded-[3rem] border border-parea-black transition-colors focus:outline-none label';

  const variantStyles = {
    primary: 'text-parea-black',
    secondary: 'bg-parea-white text-parea-black hover:bg-parea-grey',
    tertiary: 'bg-transparent text-parea-black border-none underline',
  };

  const sizeStyles = {
    sm: 'px-3 py-1.5 text-small',
    md: 'px-4 py-2 text-regular',
    lg: 'px-6 py-3 text-regular h-[2.9375rem]',
  };

  const disabledStyles = 'opacity-50 cursor-not-allowed';

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className={`
        ${baseStyles}
        ${variantStyles[variant]}
        ${sizeStyles[size]}
        ${disabled ? disabledStyles : ''}
        ${className}
      `}
      style={{
        fontFamily: 'var(--font-ibm-plex-mono), monospace',
        fontWeight: 500,
        ...(variant === 'primary' ? { backgroundColor: '#DDFF30' } : {}),
        transition: 'background-color 0.2s ease',
      }}
      onMouseEnter={(e) => {
        if (variant === 'primary' && !disabled) {
          e.currentTarget.style.backgroundColor = '#C8E82A'; // Slightly darker yellow
        }
      }}
      onMouseLeave={(e) => {
        if (variant === 'primary' && !disabled) {
          e.currentTarget.style.backgroundColor = '#DDFF30';
        }
      }}
    >
      {children}
    </button>
  );
}
