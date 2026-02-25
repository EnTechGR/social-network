'use client';

import Link from 'next/link';
import { useEffect, useMemo, useState } from 'react';
import {
  getPublicUsersForSearch,
  type ForumUser,
  getAllPublicPostsForSearch,
  getAllGroups,
  getMemberEventsForSearch,
  type SearchEventItem,
} from '@/lib/api';
import Tabs from './Tabs';
import { MessageCircle } from 'lucide-react';

const SEARCH_TABS = ['Posts', 'Events', 'Users', 'Groups'];

interface FeedPost {
  id: string;
  created_at: string;
  visibility: string;
  author_id: string;
  author_nickname: string;
  author_first_name: string;
  author_last_name: string;
  author_avatar_url: string;
  author_avatar_thumb_url: string;
  title: string;
  content: string;
  images: Array<{
    image_id: string;
    url: string;
    thumbnail_url: string;
    display_order: number;
  }>;
  like_count: number;
  dislike_count: number;
  comment_count: number;
  viewer_reaction: number | null;
}

interface Group {
  id: string;
  owner_id: string;
  owner_nickname: string;
  title: string;
  description?: string | null;
  member_count: number;
  created_at: string;
}

interface SearchSuggestionsProps {
  query?: string;
  onClose?: () => void;
}

function formatEventDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString('en-GB', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function GroupSuggestion({
  group,
  isLast,
  onClose,
}: {
  group: Group;
  isLast: boolean;
  onClose?: () => void;
}) {
  const descriptionPreview = group.description 
    ? (group.description.length > 80 ? group.description.substring(0, 80) + '...' : group.description)
    : 'No description';

  return (
    <Link
      href={`/group/${group.id}`}
      onClick={onClose}
      className={[
        'flex w-full flex-col gap-2 bg-parea-white px-6 py-4 no-underline transition-colors duration-200 hover:bg-parea-yellow',
        isLast ? '' : 'border-b border-parea-border',
      ].join(' ')}
    >
      <span className="font-sans text-base font-semibold text-parea-black">{group.title}</span>
      <p className="font-sans text-sm text-parea-black/70">{descriptionPreview}</p>
      <div className="flex items-center gap-4 font-mono text-xs text-parea-black/50">
        <span>👥 {group.member_count} {group.member_count === 1 ? 'member' : 'members'}</span>
        <span>• Owner: @{group.owner_nickname}</span>
      </div>
    </Link>
  );
}

function PostSuggestion({
  post,
  isLast,
  onClose,
}: {
  post: FeedPost;
  isLast: boolean;
  onClose?: () => void;
}) {
  const authorName = [post.author_first_name, post.author_last_name].filter(Boolean).join(' ').trim() || post.author_nickname;
  const contentPreview = post.content.length > 100 ? post.content.substring(0, 100) + '...' : post.content;

  return (
    <Link
      href={`/post/${post.id}`}
      onClick={onClose}
      className={[
        'flex w-full flex-col gap-2 bg-parea-white px-6 py-4 no-underline transition-colors duration-200 hover:bg-parea-yellow',
        isLast ? '' : 'border-b border-parea-border',
      ].join(' ')}
    >
      <div className="flex items-center gap-2">
        <span className="font-mono text-xs uppercase tracking-wide text-parea-black/60">@{post.author_nickname}</span>
        <span className="font-sans text-xs text-parea-black/50">• {authorName}</span>
      </div>
      {post.title && (
        <span className="font-sans text-base font-semibold text-parea-black">{post.title}</span>
      )}
      <p className="font-sans text-sm text-parea-black/70">{contentPreview}</p>
      <div className="flex items-center gap-4 font-mono text-xs text-parea-black/50">
        <span className="flex items-center gap-1">
          <MessageCircle className="w-4 h-4" />
          {post.comment_count}
        </span>
      </div>
    </Link>
  );
}

function EventSuggestion({
  event,
  isLast,
  onClose,
}: {
  event: SearchEventItem;
  isLast: boolean;
  onClose?: () => void;
}) {
  const descriptionPreview = event.description
    ? (event.description.length > 100 ? `${event.description.slice(0, 100)}...` : event.description)
    : 'No description';

  return (
    <Link
      href={`/event/${event.id}`}
      onClick={onClose}
      className={[
        'flex w-full flex-col gap-2 bg-parea-white px-6 py-4 no-underline transition-colors duration-200 hover:bg-parea-yellow',
        isLast ? '' : 'border-b border-parea-border',
      ].join(' ')}
    >
      <span className="font-sans text-base font-semibold text-parea-black">{event.title}</span>
      <p className="font-sans text-sm text-parea-black/70">{descriptionPreview}</p>
      <div className="flex items-center gap-4 font-mono text-xs text-parea-black/50">
        <span>📅 {formatEventDate(event.event_time)}</span>
        <span>• Group: {event.group_title || 'Group'}</span>
      </div>
    </Link>
  );
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
    </Link>
  );
}

export default function SearchSuggestions({ query = '', onClose }: SearchSuggestionsProps) {
  const [activeTab, setActiveTab] = useState('Users');
  const [users, setUsers] = useState<ForumUser[]>([]);
  const [posts, setPosts] = useState<FeedPost[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [events, setEvents] = useState<SearchEventItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;

    const loadData = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const [loadedUsers, loadedPosts, groupsData, memberEvents] = await Promise.all([
                getPublicUsersForSearch(),
          getAllPublicPostsForSearch(),
          getAllGroups(),
          getMemberEventsForSearch(),
        ]);

        if (mounted) {
          setUsers(loadedUsers);
          setPosts(loadedPosts);
          setGroups(groupsData);
          setEvents(memberEvents);
        }
      } catch (err) {
        if (mounted) {
          setError(err instanceof Error ? err.message : 'Failed to load data');
        }
      } finally {
        if (mounted) {
          setIsLoading(false);
        }
      }
    };

    loadData();

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

  const filteredPosts = useMemo(() => {
    const value = query.trim().toLowerCase();

    if (!value) {
      return posts;
    }

    return posts.filter((post) =>
      [post.title, post.content, post.author_nickname, post.author_first_name, post.author_last_name]
        .filter(Boolean)
        .some((field) => field.toLowerCase().includes(value))
    );
  }, [query, posts]);

  const filteredGroups = useMemo(() => {
    const value = query.trim().toLowerCase();

    if (!value) {
      return groups;
    }

    return groups.filter((group) =>
      [group.title, group.description, group.owner_nickname]
        .filter(Boolean)
        .some((field) => field?.toLowerCase().includes(value))
    );
  }, [query, groups]);

  const filteredEvents = useMemo(() => {
    const value = query.trim().toLowerCase();

    if (!value) {
      return events;
    }

    return events.filter((event) =>
      [event.title, event.description || '', event.group_title]
        .filter(Boolean)
        .some((field) => field.toLowerCase().includes(value))
    );
  }, [query, events]);

  return (
    <div className="w-full overflow-hidden rounded-xl border border-parea-border bg-parea-white shadow-lg">
      <div className="border-b border-parea-border px-6 py-4">
        <Tabs tabs={SEARCH_TABS} defaultTab="Users" onTabChange={setActiveTab} />
      </div>

      {/* Events Tab */}
      {activeTab === 'Events' && isLoading && (
        <div className="px-6 py-8 font-sans text-sm text-parea-black/70">Loading events...</div>
      )}

      {activeTab === 'Events' && !isLoading && error && (
        <div className="px-6 py-8 font-sans text-sm text-red-600">{error}</div>
      )}

      {activeTab === 'Events' && !isLoading && !error && filteredEvents.length === 0 && (
        <div className="px-6 py-8 font-sans text-sm text-parea-black/70">
          No events found in groups you are a member of.
        </div>
      )}

      {activeTab === 'Events' && !isLoading && !error && filteredEvents.length > 0 && (
        <div className="max-h-[420px] overflow-y-auto">
          {filteredEvents.map((event, index) => (
            <EventSuggestion
              key={event.id}
              event={event}
              isLast={index === filteredEvents.length - 1}
              onClose={onClose}
            />
          ))}
        </div>
      )}

      {/* Posts Tab */}
      {activeTab === 'Posts' && isLoading && (
        <div className="px-6 py-8 font-sans text-sm text-parea-black/70">Loading posts...</div>
      )}

      {activeTab === 'Posts' && !isLoading && error && (
        <div className="px-6 py-8 font-sans text-sm text-red-600">{error}</div>
      )}

      {activeTab === 'Posts' && !isLoading && !error && filteredPosts.length === 0 && (
        <div className="px-6 py-8 font-sans text-sm text-parea-black/70">No posts found.</div>
      )}

      {activeTab === 'Posts' && !isLoading && !error && filteredPosts.length > 0 && (
        <div className="max-h-[420px] overflow-y-auto">
          {filteredPosts.map((post, index) => (
            <PostSuggestion
              key={post.id}
              post={post}
              isLast={index === filteredPosts.length - 1}
              onClose={onClose}
            />
          ))}
        </div>
      )}

      {/* Groups Tab */}
      {activeTab === 'Groups' && isLoading && (
        <div className="px-6 py-8 font-sans text-sm text-parea-black/70">Loading groups...</div>
      )}

      {activeTab === 'Groups' && !isLoading && error && (
        <div className="px-6 py-8 font-sans text-sm text-red-600">{error}</div>
      )}

      {activeTab === 'Groups' && !isLoading && !error && filteredGroups.length === 0 && (
        <div className="px-6 py-8 font-sans text-sm text-parea-black/70">No groups found.</div>
      )}

      {activeTab === 'Groups' && !isLoading && !error && filteredGroups.length > 0 && (
        <div className="max-h-[420px] overflow-y-auto">
          {filteredGroups.map((group, index) => (
            <GroupSuggestion
              key={group.id}
              group={group}
              isLast={index === filteredGroups.length - 1}
              onClose={onClose}
            />
          ))}
        </div>
      )}

      {/* Users Tab */}
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
