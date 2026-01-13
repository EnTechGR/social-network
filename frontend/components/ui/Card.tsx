/**
 * components/ui/Card.tsx
 *
 * Reusable card component for displaying content in a contained box.
 * Common use cases: user profiles, posts, settings panels, etc.
 */

interface CardProps {
  children: React.ReactNode;
  className?: string;
}

export default function Card({ children, className = '' }: CardProps) {
  return (
    <div
      className={`
        bg-white dark:bg-gray-800
        border border-gray-200 dark:border-gray-700
        rounded-lg shadow-sm
        p-6
        ${className}
      `}
    >
      {children}
    </div>
  );
}

/**
 * Card.Header - Optional header section for the card
 */
Card.Header = function CardHeader({ children, className = '' }: CardProps) {
  return (
    <div className={`mb-4 pb-4 border-b border-gray-200 dark:border-gray-700 ${className}`}>
      {children}
    </div>
  );
};

/**
 * Card.Title - Title for the card
 */
Card.Title = function CardTitle({ children, className = '' }: CardProps) {
  return (
    <h3 className={`text-lg font-semibold text-gray-900 dark:text-white ${className}`}>
      {children}
    </h3>
  );
};

/**
 * Card.Content - Main content area
 */
Card.Content = function CardContent({ children, className = '' }: CardProps) {
  return (
    <div className={`text-gray-600 dark:text-gray-300 ${className}`}>
      {children}
    </div>
  );
};
