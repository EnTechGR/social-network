'use client';

import Link from 'next/link';
import { useEffect, useMemo, useState } from 'react';
import { getForumUsers, type ForumUser } from '@/lib/api';
import Tabs from './Tabs';

const SEARCH_TABS = ['Posts', 'Events', 'Users', 'Groups'];

interface SearchSuggestionsProps {
  query?: string;
  onClose?: () => void;
}

function UserSuggestion({
  user,
  isLast,
  onClose,
}: {
  user: ForumUser;
  isLast: boolean;
  onClose?: () => void;
}) {
  const fullName = [user.first_name, user.last_name].filter(Boolean).join(' ').trim() || user.nickname;

  return (
    <Link
      href={`/profile/${user.id}`}
      onClick={onClose}
      className={[
        'flex w-full items-center justify-between gap-3 bg-parea-white px-6 py-4 no-underline transition-colors duration-200 hover:bg-parea-yellow',
        isLast ? '' : 'border-b border-parea-border',
      ].join(' ')}
    >
      <div className="flex min-w-0 flex-col">
        <span className="truncate font-sans text-base font-semibold text-parea-black">{fullName}</span>
        <span className="truncate font-mono text-xs uppercase tracking-wide text-parea-black/60">@{user.nickname}</span>
      </div>
      <span
        className={[
          'rounded-full border px-2 py-1 font-mono text-[10px] uppercase tracking-wide',
          user.is_online ? 'border-green-700 text-green-700' : 'border-parea-border text-parea-black/60',
        ].join(' ')}
      >
        {user.is_online ? 'Online' : 'Offline'}
      </span>
    </Link>
  );
}

export default function SearchSuggestions({ query = '', onClose }: SearchSuggestionsProps) {
  const [activeTab, setActiveTab] = useState('Users');
  const [users, setUsers] = useState<ForumUser[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;

    const loadUsers = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const loadedUsers = await getForumUsers();
        if (mounted) {
          setUsers(loadedUsers);
        }
      } catch (err) {
        if (mounted) {
          setError(err instanceof Error ? err.message : 'Failed to load users');
        }
      } finally {
        if (mounted) {
          setIsLoading(false);
        }
      }
    };

    loadUsers();

    return () => {
      mounted = false;
    };
  }, []);

  const filteredUsers = useMemo(() => {
    const value = query.trim().toLowerCase();

    if (!value) {
      return users;
    }

    return users.filter((user) =>
      [user.nickname, user.first_name, user.last_name, user.email]
        .filter(Boolean)
        .some((field) => field.toLowerCase().includes(value))
    );
  }, [query, users]);

  return (
    <div className="w-full overflow-hidden rounded-xl border border-parea-border bg-parea-white shadow-lg">
      <div className="border-b border-parea-border px-6 py-4">
        <Tabs tabs={SEARCH_TABS} defaultTab="Users" onTabChange={setActiveTab} />
      </div>

      {activeTab !== 'Users' && (
        <div className="px-6 py-8 font-sans text-sm text-parea-black/70">
          The Users tab is wired to live forum users. Other tabs are not connected yet.
        </div>
      )}

      {activeTab === 'Users' && isLoading && (
        <div className="px-6 py-8 font-sans text-sm text-parea-black/70">Loading users...</div>
      )}

      {activeTab === 'Users' && !isLoading && error && (
        <div className="px-6 py-8 font-sans text-sm text-red-600">{error}</div>
      )}

      {activeTab === 'Users' && !isLoading && !error && filteredUsers.length === 0 && (
        <div className="px-6 py-8 font-sans text-sm text-parea-black/70">No users found.</div>
      )}

      {activeTab === 'Users' && !isLoading && !error && filteredUsers.length > 0 && (
        <div className="max-h-[420px] overflow-y-auto">
          {filteredUsers.map((user, index) => (
            <UserSuggestion
              key={user.id}
              user={user}
              isLast={index === filteredUsers.length - 1}
              onClose={onClose}
            />
          ))}
        </div>
      )}
    </div>
  );
}
