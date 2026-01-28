/**
 * lib/api.ts
 *
 * API client functions for making HTTP requests.
 * These functions abstract away fetch calls and provide type-safe interfaces.
 */

import type { AuthResponse, ApiErrorResponse, LoginRequest, RegisterStep1Request, RegisterStep2Request, VerifyResponse } from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

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
