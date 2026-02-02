// components/ui/SidebarToggle.tsx
export default function SidebarToggle({ isOpen, onClick }: { isOpen: boolean; onClick: () => void }) {
    return (
      <button
        onClick={onClick}
        className="
          w-9 h-9 
          rounded-full 
          bg-transparent 
          border border-parea-yellow
          flex items-center justify-center
          hover:bg-parea-yellow/10
          transition-colors
          cursor-pointer
        "
        aria-label={isOpen ? 'Close sidebar' : 'Open sidebar'}
      >
        <svg
          width="15"
          height="15"
          viewBox="0 0 15 15"
          fill="none"
          className={`transition-transform ${isOpen ? 'rotate-180' : ''}`}
        >
          <path
            d="M0.5 7.5H14.5M14.5 7.5L7.5 0.5M14.5 7.5L7.5 14.5"
            stroke="var(--parea-yellow)"
            strokeWidth="1"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </button>
    );
  }