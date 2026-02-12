/**
 * lib/utils.ts
 *
 * General utility functions that can be used throughout the app.
 * These are pure functions - they don't have side effects and always
 * return the same output for the same input.
 */

/**
 * Format a date to a readable string
 * Example: formatDate(new Date()) => "January 8, 2026"
 */
export function formatDate(date: Date): string {
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }).format(date);
}

/**
 * Format a relative time string
 * Example: getRelativeTime(new Date(Date.now() - 3600000)) => "1 hour ago"
 */
export function getRelativeTime(date: Date): string {
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffInSeconds < 60) return 'just now';
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)} minutes ago`;
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)} hours ago`;
  return `${Math.floor(diffInSeconds / 86400)} days ago`;
}

/**
 * Truncate text to a maximum length
 * Example: truncate("Hello World", 5) => "Hello..."
 */
export function truncate(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength) + '...';
}

/**
 * Generate initials from a name
 * Example: getInitials("John Doe") => "JD"
 */
export function getInitials(name: string): string {
  return name
    .split(' ')
    .map(word => word[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}
