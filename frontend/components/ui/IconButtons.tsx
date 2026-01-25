/**
 * components/ui/IconButton.tsx
 *
 * Parea Icon Button Component
 * For: arrow (left/right) and close buttons
 * Can be used with or without text
 * Fixed size: 2.25rem x 2.25rem (36px) for icon-only, auto for with text
 */

// Arrow icon component (inline)
function ArrowIcon({ direction = 'right', className = '' }: { direction?: 'left' | 'right'; className?: string }) {
  const baseRotation = direction === 'left' ? 'rotate-180' : '';
  const hoverRotation = direction === 'left' ? 'group-hover:rotate-135' : 'group-hover:-rotate-45';

  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="15"
      height="15"
      viewBox="0 0 15 15"
      fill="none"
      className={`w-3.5 h-3.5 shrink-0 transition-transform duration-300 ease-in-out ${baseRotation} ${hoverRotation} ${className}`}
    >
      <path
        d="M0.5 7.5H14.5M14.5 7.5L7.5 0.5M14.5 7.5L7.5 14.5"
        stroke="black"
        strokeWidth="1"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

interface IconButtonProps {
  variant: 'arrow-left' | 'arrow-right' | 'close';
  onClick?: (e?: React.MouseEvent<HTMLButtonElement>) => void;
  disabled?: boolean;
  className?: string;
  'aria-label': string;
  text?: string; // Optional text to display next to the icon
  type?: 'button' | 'submit' | 'reset';
  size?: 'sm' | 'md'; // sm = 24px, md = 36px (default)
  transparent?: boolean; // If true, no background color
}

function CloseIcon({ size = 'md' }: { size?: 'sm' | 'md' }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="12"
      height="12"
      viewBox="0 0 12 12"
      fill="none"
      className={`${size === 'sm' ? 'w-2.5 h-2.5' : 'w-3.5 h-3.5'} shrink-0`}
    >
      <path
        d="M11.5 0.5L0.5 11.5M0.5 0.5L11.5 11.5"
        stroke="black"
        strokeWidth="1"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export default function IconButton({
  variant,
  onClick,
  disabled = false,
  className = '',
  'aria-label': ariaLabel,
  text,
  type = 'button',
  size = 'md',
  transparent = false,
}: IconButtonProps) {
  const sizeStyles = size === 'sm' ? 'w-6 h-6 p-1.5' : 'w-[2.25rem] h-[2.25rem] p-3';

  const iconButtonStyles = `
    group
    flex
    justify-center
    items-center
    ${sizeStyles}
    rounded-[6.25rem]
    border
    border-parea-black
    text-parea-black
  `;

  const withTextStyles = `
    group
    flex
    items-center
    gap-2
    px-0
    py-0
    rounded-none
    border-none
    text-parea-black
    label
  `;

  const baseStyles = text ? withTextStyles : iconButtonStyles;
  const disabledStyles = disabled ? 'opacity-50 cursor-not-allowed' : '';

  // Determine which icon to show
  const icon =
    variant === 'arrow-left' ? (
      <ArrowIcon direction="left" />
    ) : variant === 'arrow-right' ? (
      <ArrowIcon direction="right" />
    ) : (
      <CloseIcon size={size} />
    );

  // If text is provided, render icon + text (no background on button, only on icon circle)
  // If no text, render just the icon button (with background)
  if (text) {
    return (
      <button
        type={type}
        onClick={onClick}
        disabled={disabled}
        aria-label={ariaLabel}
        className={`
          ${baseStyles}
          ${disabledStyles}
          ${className}
        `}
        style={{
          fontFamily: 'var(--font-ibm-plex-mono), monospace',
          fontWeight: 500,
        }}
      >
        {variant === 'arrow-left' && (
          <div style={{ backgroundColor: '#DDFF30' }} className="w-9 h-9 rounded-[6.25rem] border border-parea-black flex items-center justify-center">
            {icon}
          </div>
        )}
        <span>{text}</span>
        {(variant === 'arrow-right' || variant === 'close') && (
          <div style={{ backgroundColor: '#DDFF30' }} className="w-9 h-9 rounded-[6.25rem] border border-parea-black flex items-center justify-center">
            {icon}
          </div>
        )}
      </button>
    );
  }

  // Icon-only button
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      aria-label={ariaLabel}
      className={`
        ${baseStyles}
        ${disabledStyles}
        ${className}
      `}
      style={transparent ? undefined : { backgroundColor: '#DDFF30' }}
    >
      {icon}
    </button>
  );
}
