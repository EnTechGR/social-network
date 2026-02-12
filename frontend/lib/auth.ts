/**
 * lib/auth.ts
 *
 * Authentication utility functions for checking user session and managing auth state
 */

import { verifyAuth } from './api';
import type { VerifyResponse } from '@/types';

/**
 * Verify the current session with the API
 * @returns VerifyResponse if authenticated, null otherwise
 */
export async function verifySession(): Promise<VerifyResponse | null> {
  try {
    const response = await verifyAuth();
    // Store CSRF token and user data (actual auth is via HttpOnly cookie)
    if (typeof window !== 'undefined') {
      localStorage.setItem('csrf_token', response.csrf_token);
      localStorage.setItem('user_data', JSON.stringify(response.user));
    }
    return response;
  } catch (error) {
    // Session is invalid - clear auth data
    clearAuth();
    return null;
  }
}

/**
 * Check if a user appears to be authenticated (has user data)
 * Note: This is a client-side check only. Actual auth is via HttpOnly cookie.
 * Use verifySession() for server-validated authentication.
 * @returns true if user data exists, false otherwise
 */
export function isAuthenticated(): boolean {
  if (typeof window === 'undefined') {
    // Server-side rendering - can't access localStorage
    return false;
  }
  
  const userData = localStorage.getItem('user_data');
  return !!userData;
}

/**
 * Get the CSRF token for state-changing requests
 * @returns the CSRF token or null if not available
 */
export function getCsrfToken(): string | null {
  if (typeof window === 'undefined') {
    return null;
  }
  
  return localStorage.getItem('csrf_token');
}

/**
 * Get the current user's data from localStorage
 * Note: This is for UI purposes only, not for authentication
 * @returns the user data or null if not available
 */
export function getUserData(): any | null {
  if (typeof window === 'undefined') {
    return null;
  }
  
  const userData = localStorage.getItem('user_data');
  return userData ? JSON.parse(userData) : null;
}

/**
 * Get the current user's ID (convenience function)
 * @returns the user ID or null if not available
 */
export function getUserId(): string | null {
  const userData = getUserData();
  return userData?.id || null;
}

/**
 * Clear authentication data (logout)
 */
export function clearAuth(): void {
  if (typeof window === 'undefined') {
    return;
  }
  
  localStorage.removeItem('csrf_token');
  localStorage.removeItem('user_data');
}
