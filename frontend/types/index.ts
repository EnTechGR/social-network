/**
 * types/index.ts
 *
 * Shared TypeScript type definitions used across the app.
 * Centralizing types here makes them reusable and maintains consistency.
 */

/**
 * User type definition
 */
export interface User {
  id: string;
  name: string;
  email: string;
  username?: string;
  bio?: string;
  avatar?: string;
  createdAt?: Date;
}

/**
 * Post type definition
 */
export interface Post {
  id: string;
  userId: string;
  content: string;
  imageUrl?: string;
  likes: number;
  comments: number;
  createdAt: Date;
  updatedAt?: Date;
}

/**
 * Comment type definition
 */
export interface Comment {
  id: string;
  postId: string;
  userId: string;
  content: string;
  createdAt: Date;
}

/**
 * API Response wrapper
 */
export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

/**
 * Pagination parameters
 */
export interface PaginationParams {
  page: number;
  limit: number;
}

/**
 * Paginated response
 */
export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  limit: number;
  total: number;
  hasMore: boolean;
}

/**
 * Auth User type (matches API response)
 */
export interface AuthUser {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  nickname?: string;
  date_of_birth: string;
  gender: string;
  about_me?: string;
  is_private: boolean;
  created_at: string;
}

/**
 * Auth Response type
 */
export interface AuthResponse {
  id: string;
  csrf_token: string;
  user: AuthUser;
}

/**
 * API Error Response
 */
export interface ApiErrorResponse {
  code: number;
  error: string;
  message: string;
}

/**
 * Login Request
 */
export interface LoginRequest {
  login: string;
  password: string;
}

/**
 * Register Request (Step 1)
 */
export interface RegisterStep1Request {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  date_of_birth: string;
  gender: 'male' | 'female' | 'other' | 'prefer_not_to_say';
  avatar?: File;

  // Optional profile fields that can be sent during registration
  nickname?: string;
  about_me?: string;
  is_private?: boolean;
}

/**
 * Register Request (Step 2)
 */
export interface RegisterStep2Request {
  nickname: string;
  about_me?: string;
  is_private?: boolean;
}
