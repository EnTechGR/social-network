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
    <div className="flex items-start gap-4 p-4 border border-parea-black rounded-lg bg-parea-white">
      {/* User Avatar */}
      <div className="w-12 h-12 rounded-full bg-parea-black text-parea-white flex items-center justify-center font-semibold shrink-0">
        {getInitials(user.name)}
      </div>

      {/* User Info */}
      <div className="flex-1">
        <h3 className="font-semibold text-parea-black">{user.name}</h3>
        <p className="text-sm text-parea-black/60">{user.email}</p>
        {user.bio && (
          <p className="mt-2 text-sm text-parea-black/80">{user.bio}</p>
        )}
      </div>

      {/* Action Button */}
      {onFollow && (
        <Button variant="primary" size="sm" onClick={() => onFollow(user.id)}>
          Follow
        </Button>
      )}
    </div>
  );
}
