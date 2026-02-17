/**
 * types/feed.ts
 * 
 * TypeScript types for feed-related API responses
 */

export interface FeedPost {
  id: string;
  user_id: string;
  nickname: string;
  first_name: string;
  last_name: string;
  avatar_url?: string | null;
  group_id?: string | null;
  group_name?: string | null;
  visibility: 'public' | 'followers' | 'private';
  title?: string | null;
  content?: string | null;
  reaction_count: number;
  comment_count: number;
  user_has_reacted: boolean;
  user_reaction_type?: number | null;
  image_urls?: string[];
  thumbnail_urls?: string[];
  created_at: string;
  updated_at?: string | null;
}

export interface FeedPaginationMeta {
  next_cursor?: string | null;
  has_more: boolean;
  count: number;
  limit: number;
}

export interface FeedResponse {
  posts: FeedPost[];
  pagination: FeedPaginationMeta;
}

export interface FeedQueryParams {
  cursor?: string;
  limit?: number;
  sort?: 'latest' | 'popular' | 'following';
  visibility?: 'all' | 'public' | 'followers' | 'private';
  time_range?: 'all' | 'today' | 'week' | 'month';
}

// Placeholder types for other feed tabs (to be implemented)
export interface FeedGroup {
  id: string;
  title: string;
  description?: string;
  member_count: number;
  created_at: string;
}

export interface FeedUser {
  id: string;
  nickname: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
  is_private: boolean;
  is_following: boolean;
}

export interface FeedEvent {
  id: string;
  title: string;
  description?: string;
  event_time: string;
  group_id?: string;
  group_name?: string;
  creator_nickname: string;
  created_at: string;
}