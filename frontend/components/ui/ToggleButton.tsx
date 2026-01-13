/**
 * components/ui/ToggleButton.tsx
 *
 * Parea Toggle Button Component
 * A toggle/switch with on/off states
 */

interface ToggleButtonProps {
    isOn?: boolean;
    onChange?: (isOn: boolean) => void;
    disabled?: boolean;
    'aria-label': string;
    className?: string;
  }
  
  export default function ToggleButton({
    isOn = false,
    onChange,
    disabled = false,
    'aria-label': ariaLabel,
    className = '',
  }: ToggleButtonProps) {
    
    const handleClick = () => {
      if (!disabled && onChange) {
        onChange(!isOn);
      }
    };
  
    // Frame styles
    const frameStyles = `
      flex
      w-10
      h-6
      flex-col
      justify-center
      rounded-full
      border
      transition-all
      duration-300
      ease-in-out
      ${isOn 
        ? 'items-end bg-parea-black border-parea-black' 
        : 'items-start bg-parea-grey border-parea-black'
      }
      ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
    `;
  
    // Circle styles
    const circleStyles = `
      w-5
      h-5
      rounded-full
      border
      border-parea-black
      bg-parea-yellow
      transition-all
      duration-300
      ease-in-out
      ${isOn ? '' : 'opacity-90'}
    `;
  
    return (
      <button
        type="button"
        role="switch"
        aria-checked={isOn}
        aria-label={ariaLabel}
        onClick={handleClick}
        disabled={disabled}
        className={`${frameStyles} ${className}`}
      >
        <span className={circleStyles} />
      </button>
    );
  }