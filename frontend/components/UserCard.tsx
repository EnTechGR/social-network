/**
 * components/UserCard.tsx
 *
 * Displays user information in a card format.
 * Example of a component that uses other components (Card, Button).
 */

import Card from './ui/Card';
import Button from './ui/Button';
import { getInitials } from '@/lib/utils';

interface UserCardProps {
  user: {
    id: string;
    name: string;
    email: string;
    bio?: string;
  };
  onFollow?: (userId: string) => void;
}

export default function UserCard({ user, onFollow }: UserCardProps) {
  return (
    <Card>
      <div className="flex items-start gap-4">
        {/* User Avatar */}
        <div className="w-12 h-12 rounded-full bg-blue-600 text-white flex items-center justify-center font-semibold">
          {getInitials(user.name)}
        </div>

        {/* User Info */}
        <div className="flex-1">
          <h3 className="font-semibold text-gray-900 dark:text-white">
            {user.name}
          </h3>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            {user.email}
          </p>
          {user.bio && (
            <p className="mt-2 text-sm text-gray-600 dark:text-gray-300">
              {user.bio}
            </p>
          )}
        </div>

        {/* Action Button */}
        {onFollow && (
          <Button
            variant="primary"
            size="sm"
            onClick={() => onFollow(user.id)}
          >
            Follow
          </Button>
        )}
      </div>
    </Card>
  );
}
