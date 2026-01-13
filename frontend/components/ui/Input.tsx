/**
 * components/ui/Input.tsx
 *
 * Reusable input component for forms.
 * Follows Parea Design System styling.
 * Includes built-in label, error message, and validation states.
 */

interface InputProps {
  type?: 'text' | 'email' | 'password' | 'number' | 'date';
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  error?: string;
  required?: boolean;
  disabled?: boolean;
  onFocus?: () => void;
  onBlur?: () => void;
}

export default function Input({
  type = 'text',
  placeholder,
  value,
  onChange,
  error,
  required = false,
  disabled = false,
  onFocus,
  onBlur,
}: InputProps) {
  return (
    <div className="flex flex-col gap-1">
      <input
        type={type}
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onFocus={onFocus}
        onBlur={onBlur}
        disabled={disabled}
        className={`
          flex
          h-[2.9375rem]
          px-3 py-2
          items-center
          gap-2
          self-stretch
          rounded-[3rem]
          border
          border-parea-black
          bg-parea-white
          text-parea-black
          placeholder:text-[#00000099]
          placeholder:font-mono
          placeholder:text-regular
          placeholder:font-medium
          placeholder:leading-relaxed
          placeholder:uppercase
          focus:outline-none
          disabled:opacity-50 disabled:cursor-not-allowed
          ${
            error
              ? 'border-red-500'
              : ''
          }
        `}
        style={{
          fontFamily: 'var(--font-body), system-ui, sans-serif',
        }}
      />

      {error && (
        <span className="text-sm text-red-600">
          {error}
        </span>
      )}
    </div>
  );
}
