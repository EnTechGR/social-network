/**
 * lib/api.ts
 *
 * API client functions for making HTTP requests.
 * These functions abstract away fetch calls and provide type-safe interfaces.
 */

import type { AuthResponse, ApiErrorResponse, LoginRequest, RegisterStep1Request, RegisterStep2Request, VerifyResponse } from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface NotificationItem {
  id: string;
  from_user_id?: string;
  nickname: string;
  type: string;
  post_id?: string;
  comment_id?: string | null;
  created_at: string;
  read: boolean;
  visible: boolean;
}

export interface NotificationsResponse {
  count: number;
  notifications: NotificationItem[];
}

export interface ForumUser {
  id: string;
  nickname: string;
  email: string;
  first_name: string;
  last_name: string;
  date_of_birth: string;
  gender: string;
  is_online: boolean;
}

export interface FollowRelationship {
  follower_id: string;
  followee_id: string;
  status: 'pending' | 'accepted' | 'blocked';
  created_at: string;
  updated_at?: string;
}

export interface FollowRequestsResponse {
  pending: FollowRelationship[];
  accepted: FollowRelationship[];
}

/**
 * Generic fetch wrapper with error handling
 */
async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const isFormData = options?.body instanceof FormData;
  const headers = isFormData
    ? { ...options?.headers }
    : {
        'Content-Type': 'application/json',
        ...options?.headers,
      };

  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers,
      credentials: 'include', // Include cookies for session management
    });

    if (!response.ok) {
      const errorData: ApiErrorResponse = await response.json().catch(() => ({
        code: response.status,
        error: response.statusText,
        message: response.statusText,
      }));
      throw new Error(errorData.message || errorData.error || `API Error: ${response.statusText}`);
    }

    // Handle empty responses (e.g., 204 No Content or empty body)
    const contentType = response.headers.get('content-type');
    const contentLength = response.headers.get('content-length');
    
    if (contentLength === '0' || !contentType?.includes('application/json')) {
      return undefined as T;
    }

    return response.json();
  } catch (error: any) {
    // Handle network errors (connection refused, CORS, etc.)
    if (error instanceof TypeError && error.message.includes('fetch')) {
      throw new Error(`Unable to connect to the server. Please make sure the backend API is running on ${API_BASE_URL}`);
    }
    // Re-throw other errors
    throw error;
  }
}

function getCSRFToken(): string {
  if (typeof window === 'undefined') return '';
  return localStorage.getItem('csrf_token') || '';
}

/**
 * Example: Fetch all users
 * Usage in a component: const users = await getUsers();
 */
export async function getUsers() {
  return fetchAPI<{ id: string; name: string; email: string }[]>('/users');
}

/**
 * Example: Fetch a single user by ID
 * Usage: const user = await getUserById('123');
 */
export async function getUserById(id: string) {
  return fetchAPI<{ id: string; name: string; email: string }>(`/users/${id}`);
}

/**
 * Example: Create a new user
 * Usage: const newUser = await createUser({ name: 'John', email: 'john@example.com' });
 */
export async function createUser(data: { name: string; email: string }) {
  return fetchAPI<{ id: string; name: string; email: string }>('/users', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

/**
 * Login a user
 * POST /api/v1/login
 */
export async function login(credentials: { email: string; password: string }): Promise<AuthResponse> {
  return fetchAPI<AuthResponse>('/api/v1/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ login: credentials.email, password: credentials.password }),
  });
}

/**
 * Register a new user (Step 1)
 * POST /api/v1/register
 * Accepts multipart/form-data for avatar uploads
 */
export async function registerStep1(data: RegisterStep1Request): Promise<AuthResponse> {
  const formData = new FormData();
  formData.append('email', data.email);
  formData.append('password', data.password);
  formData.append('first_name', data.first_name);
  formData.append('last_name', data.last_name);
  formData.append('date_of_birth', data.date_of_birth);
  formData.append('gender', data.gender);
  
  if (data.avatar) {
    formData.append('avatar', data.avatar);
  }

  // Optional profile fields so registration can be completed in a single request
  if ((data as any).nickname) {
    formData.append('nickname', (data as any).nickname);
  }
  if ((data as any).about_me) {
    formData.append('about_me', (data as any).about_me);
  }
  if ((data as any).is_private !== undefined) {
    formData.append('is_private', (data as any).is_private ? 'true' : 'false');
  }

  return fetchAPI<AuthResponse>('/api/v1/register', {
    method: 'POST',
    body: formData,
  });
}

/**
 * Complete registration (Step 2)
 * This would typically be a PATCH or PUT to update the user profile
 * Note: You may need to adjust the endpoint based on your API
 */
export async function registerStep2(userId: string, data: RegisterStep2Request, csrfToken: string): Promise<AuthResponse> {
  const formData = new FormData();
  formData.append('nickname', data.nickname);
  if (data.about_me) {
    formData.append('about_me', data.about_me);
  }
  if (data.is_private !== undefined) {
    formData.append('is_private', data.is_private ? 'true' : 'false');
  }

  return fetchAPI<AuthResponse>(`/api/v1/users/${userId}`, {
    method: 'PATCH',
    headers: {
      'X-CSRF-Token': csrfToken,
    },
    body: formData,
  });
}

/**
 * Verify current session
 * GET /api/v1/verify
 * Returns current user and CSRF token if authenticated
 */
export async function verifyAuth(): Promise<VerifyResponse> {
  return fetchAPI<VerifyResponse>('/api/v1/verify', {
    method: 'GET',
  });
}

/**
 * Logout current user
 * POST /api/v1/logout
 * Clears session and logs out the user
 */
export async function logout(): Promise<void> {
  return fetchAPI<void>('/api/v1/logout', {
    method: 'POST',
  });
}

/**
 * Get user profile data
 * GET /api/v1/user/profile
 * Returns profile information for the current user
 */
export async function getProfile(): Promise<any> {
  return fetchAPI<any>('/api/v1/user/profile', {
    method: 'GET',
  });
}

/**
 * Get notifications for the current user.
 * GET /api/v1/notifications
 */
export async function getNotifications(): Promise<NotificationsResponse> {
  return fetchAPI<NotificationsResponse>('/api/v1/notifications', {
    method: 'GET',
  });
}

/**
 * Hide (delete) a notification for the current user.
 * DELETE /api/v1/notifications/delete/{id}
 */
export async function hideNotification(notificationId: string): Promise<{ status: string }> {
  return fetchAPI<{ status: string }>(`/api/v1/notifications/delete/${notificationId}`, {
    method: 'DELETE',
    headers: {
      'X-CSRF-Token': getCSRFToken(),
    },
  });
}

/**
 * Follow a user (private profiles become pending requests; public profiles become accepted follows).
 * POST /api/v1/follow
 */
export async function followUser(followeeId: string): Promise<{
  follower_id: string;
  followee_id: string;
  status: 'pending' | 'accepted' | 'blocked';
  created_at: string;
  updated_at?: string;
}> {
  return fetchAPI<{
    follower_id: string;
    followee_id: string;
    status: 'pending' | 'accepted' | 'blocked';
    created_at: string;
    updated_at?: string;
  }>('/api/v1/follow', {
    method: 'POST',
    headers: {
      'X-CSRF-Token': getCSRFToken(),
    },
    body: JSON.stringify({ followee_id: followeeId }),
  });
}

/**
 * Get follow requests for the current user.
 * GET /api/v1/follow/requests
 */
export async function getFollowRequests(): Promise<FollowRequestsResponse> {
  const response = await fetchAPI<Partial<FollowRequestsResponse>>('/api/v1/follow/requests', {
    method: 'GET',
  });

  return {
    pending: response.pending ?? [],
    accepted: response.accepted ?? [],
  };
}

/**
 * Get pending follow requests for the current user.
 * GET /api/v1/follow/requests/pending
 */
export async function getPendingFollowRequests(): Promise<FollowRelationship[]> {
  const response = await fetchAPI<{ pending?: FollowRelationship[] }>('/api/v1/follow/requests/pending', {
    method: 'GET',
  });

  return response.pending ?? [];
}

/**
 * Accept an incoming follow request.
 * PUT /api/v1/follow/accept/{followerId}
 */
export async function acceptFollowRequest(followerId: string): Promise<{ status: string }> {
  return fetchAPI<{ status: string }>(`/api/v1/follow/accept/${followerId}`, {
    method: 'PUT',
    headers: {
      'X-CSRF-Token': getCSRFToken(),
    },
  });
}

/**
 * Unfollow a user.
 * DELETE /api/v1/follower/delete/{followeeId}
 */
export async function unfollowUser(followeeId: string): Promise<{ status: string }> {
  return fetchAPI<{ status: string }>(`/api/v1/follower/delete/${followeeId}`, {
    method: 'DELETE',
    headers: {
      'X-CSRF-Token': getCSRFToken(),
    },
  });
}

/**
 * Remove a follower from your profile.
 * DELETE /api/v1/followee/delete/{followerId}
 */
export async function removeFollower(followerId: string): Promise<{ status: string }> {
  return fetchAPI<{ status: string }>(`/api/v1/followee/delete/${followerId}`, {
    method: 'DELETE',
    headers: {
      'X-CSRF-Token': getCSRFToken(),
    },
  });
}

/**
 * Decline an incoming follow request.
 * Alias for removeFollower when request status is pending.
 */
export async function declineFollowRequest(followerId: string): Promise<{ status: string }> {
  return removeFollower(followerId);
}

/**
 * Get all forum users (excluding current user).
 * GET /api/v1/chat/users
 */
export async function getForumUsers(): Promise<ForumUser[]> {
  const response = await fetchAPI<{ users?: ForumUser[] }>('/api/v1/chat/users', {
    method: 'GET',
  });

  return Array.isArray(response?.users) ? response.users : [];
}

/**
 * Build avatar URL from backend path (file_path or thumbnail_path).
 * Uses API base URL so static assets resolve correctly.
 */
export function getAvatarUrl(avatarPath: string | undefined): string {
  if (!avatarPath) return '/user-avatar-default.png';
  const filename = avatarPath.replace(/^uploads\//, '');
  const base = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  return `${base}/static/${filename}`;
}

/**
 * Get current user's posts (includes comments; can 500 if comment load fails).
 * GET /api/v1/user/posts
 */
export async function getMyPosts(): Promise<any[]> {
  return fetchAPI<any[]>('/api/v1/user/posts', { method: 'GET' });
}

/**
 * Get posts by user ID (no comments loaded; use for profile to avoid "Failed to load comments" 500).
 * GET /api/v1/users/posts/{userId}
 */
export async function getPostsByUserId(userId: string): Promise<any[]> {
  return fetchAPI<any[]>(`/api/v1/users/posts/${userId}`, { method: 'GET' });
}

/**
 * Get a single post by post ID (for post detail page).
 * GET /api/v1/posts/{postId}
 */
export async function getPostById(postId: string): Promise<any> {
  return fetchAPI<any>(`/api/v1/posts/${postId}`, { method: 'GET' });
}

/**
 * Get feed: all posts visible to the current user (public, followers where applicable, private only if allowed).
 * GET /api/v1/feed
 */
export async function getFeed(): Promise<any[]> {
  return fetchAPI<any[]>('/api/v1/feed', { method: 'GET' });
}

/**
 * Create a new comment on a post
 * POST /api/v1/comments/create
 */
export async function createComment(postId: string, content: string): Promise<any> {
  const csrfToken = typeof window !== 'undefined' ? localStorage.getItem('csrf_token') : null;
  
  return fetchAPI<any>('/api/v1/comments/create', {
    method: 'POST',
    headers: {
      'X-CSRF-Token': csrfToken || '',
    },
    body: JSON.stringify({
      post_id: postId,
      content: content,
    }),
  });
}

/**
 * Get comments for a post by post ID.
 * GET /api/v1/posts/{postId}/comments
 */
export async function getCommentsByPostId(postId: string): Promise<any[]> {
  return fetchAPI<any[]>(`/api/v1/posts/${postId}/comments`, { method: 'GET' });
}

/**
 * Upload user avatar
 * POST /api/v1/user/avatar
 * Accepts multipart/form-data with avatar file (max 5MB)
 */
export async function uploadAvatar(file: File): Promise<any> {
  const csrfToken = typeof window !== 'undefined' ? localStorage.getItem('csrf_token') : null;
  
  const formData = new FormData();
  formData.append('avatar', file);

  return fetchAPI<any>('/api/v1/user/avatar', {
    method: 'POST',
    headers: {
      'X-CSRF-Token': csrfToken || '',
    },
    body: formData,
  });
}

/**
 * Get another user's profile by ID
 * GET /api/v1/users/{id}
 * Returns full profile when allowed; 403 when profile is private and viewer is not a follower.
 */
export async function getUserProfile(userId: string): Promise<{
  privateProfile: true;
  message: string;
  nickname?: string;
  first_name?: string;
  last_name?: string;
} | {
  privateProfile: false;
  user: any;
  posts: any[];
  followers: any[];
  following: any[];
  counts: { posts: number; followers: number; following: number };
  is_own_profile: boolean;
}> {
  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/users/${userId}`, {
    method: 'GET',
    credentials: 'include',
  });

  const data = await response.json().catch(() => ({}));

  if (response.status === 403) {
    return {
      privateProfile: true,
      message: (data as any).message || 'This profile is private. Follow to see their content.',
      nickname: (data as any).nickname,
      first_name: (data as any).first_name,
      last_name: (data as any).last_name,
    };
  }

  if (!response.ok) {
    throw new Error((data as any).message || (data as any).error || response.statusText || 'Failed to load profile');
  }

  return {
    privateProfile: false,
    user: data.user,
    posts: data.posts ?? [],
    followers: data.followers ?? [],
    following: data.following ?? [],
    counts: data.counts ?? { posts: 0, followers: 0, following: 0 },
    is_own_profile: data.is_own_profile ?? false,
  };
}

/**
 * Update user privacy setting
 * PUT /api/v1/user/privacy
 * Toggles profile between public and private
 */
export async function updatePrivacy(isPrivate: boolean): Promise<any> {
  const csrfToken = typeof window !== 'undefined' ? localStorage.getItem('csrf_token') : null;
  
  return fetchAPI<any>('/api/v1/user/privacy', {
    method: 'PUT',
    headers: {
      'X-CSRF-Token': csrfToken || '',
    },
    body: JSON.stringify({ is_private: isPrivate }),
  });
}

/**
 * Create a new post
 * POST /api/v1/posts/create
 * Accepts multipart/form-data for image uploads
 */
export async function createPost(data: {
  title: string;
  content: string;
  visibility: 'PUBLIC' | 'FOLLOWERS' | 'PRIVATE';
  image?: File;
  allowedUserIds?: string[];
}): Promise<any> {
  const csrfToken = typeof window !== 'string' ? localStorage.getItem('csrf_token') : null;
  
  const formData = new FormData();
  formData.append('title', data.title);
  formData.append('content', data.content);
  formData.append('visibility', data.visibility.toLowerCase());
  
  if (data.image) {
    formData.append('image', data.image);
  }

  // For private posts with specific allowed users
  if (data.allowedUserIds && data.allowedUserIds.length > 0) {
    // Backend expects allowed_user_ids as JSON field, but we're using multipart
    // May need to send as JSON instead or handle it differently based on backend implementation
    formData.append('allowed_user_ids', JSON.stringify(data.allowedUserIds));
  }

  return fetchAPI<any>('/api/v1/posts/create', {
    method: 'POST',
    headers: {
      'X-CSRF-Token': csrfToken || '',
    },
    body: formData,
  });
}

// ---------------------------------------------------------------------------
// Group APIs (no backend changes)
// ---------------------------------------------------------------------------

export async function createGroup(data: { title: string; description?: string }): Promise<{ group: { id: string; title: string; description?: string; owner_nickname: string; member_count: number; created_at: string } }> {
  const csrfToken = typeof window !== 'undefined' ? localStorage.getItem('csrf_token') : null;
  return fetchAPI<any>('/api/v1/groups/create', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken || '' },
    body: JSON.stringify({ title: data.title, description: data.description || '' }),
  });
}

export async function getGroupById(groupId: string): Promise<any> {
  return fetchAPI<any>(`/api/v1/groups/${groupId}`, { method: 'GET' });
}

export async function getAllGroups(): Promise<any[]> {
  const res = await fetchAPI<any>('/api/v1/groups', { method: 'GET' });
  return Array.isArray(res) ? res : (res?.groups ?? []);
}

export async function getMyGroups(): Promise<any[]> {
  const res = await fetchAPI<any>('/api/v1/groups/my-groups', { method: 'GET' });
  return Array.isArray(res) ? res : (res?.groups ?? []);
}

export async function getGroupMembers(groupId: string): Promise<any[]> {
  const res = await fetchAPI<any>(`/api/v1/groups/members/${groupId}`, { method: 'GET' });
  return Array.isArray(res) ? res : (res?.members ?? []);
}

export async function getGroupPosts(groupId: string): Promise<any[]> {
  return fetchAPI<any[]>(`/api/v1/groups/posts/${groupId}`, { method: 'GET' });
}

export async function createGroupPost(groupId: string, data: { title: string; content: string; image?: File }): Promise<any> {
  const csrfToken = typeof window !== 'undefined' ? localStorage.getItem('csrf_token') : null;
  const formData = new FormData();
  formData.append('title', data.title);
  formData.append('content', data.content);
  if (data.image) formData.append('image', data.image);
  return fetchAPI<any>(`/api/v1/groups/posts/create/${groupId}`, {
    method: 'POST',
    headers: { 'X-CSRF-Token': csrfToken || '' },
    body: formData,
  });
}

export async function getGroupEvents(groupId: string): Promise<any[]> {
  const res = await fetchAPI<any>(`/api/v1/groups/events/${groupId}`, { method: 'GET' });
  return Array.isArray(res) ? res : (res?.events ?? []);
}

export async function createGroupEvent(groupId: string, data: { title: string; description: string; event_time: string }): Promise<any> {
  const csrfToken = typeof window !== 'undefined' ? localStorage.getItem('csrf_token') : null;
  return fetchAPI<any>(`/api/v1/groups/events/create/${groupId}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken || '' },
    body: JSON.stringify(data),
  });
}
