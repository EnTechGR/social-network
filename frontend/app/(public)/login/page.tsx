'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import AuthLayout from '@/components/auth/AuthLayout';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';
import { login } from '@/lib/api';
import { isValidEmail } from '@/lib/validations';
import Link from 'next/link';

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // Validation
    if (!email || !password) {
      setError('Please fill in all fields');
      return;
    }

    if (!isValidEmail(email)) {
      setError('Please enter a valid email address');
      return;
    }

    setIsLoading(true);

    try {
      const response = await login({ email, password });
      // Store CSRF token and user data (actual auth is via HttpOnly cookie)
      if (typeof window !== 'undefined') {
        localStorage.setItem('csrf_token', response.csrf_token);
        localStorage.setItem('user_data', JSON.stringify(response.user));
      }
      router.push('/feed');
    } catch (err: any) {
      setError(err.message || 'Invalid email or password. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthLayout>
      <div className="flex flex-col items-center justify-center w-full px-8 py-4">
        <div className="w-full max-w-md">
          <div className="flex flex-col items-center gap-4 mb-8">
            <h2 className="text-4xl font-bold text-parea-black">
              Log In
            </h2>
            <p className="text-regular text-parea-black">
              Come hang. This is your parea.
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              type="email"
              placeholder="EMAIL"
              value={email}
              onChange={setEmail}
              required
            />

            <Input
              type="password"
              placeholder="PASSWORD"
              value={password}
              onChange={setPassword}
              required
            />

            {error && (
              <div className="text-sm text-red-600 text-center">{error}</div>
            )}

            <Button
              type="submit"
              variant="primary"
              disabled={isLoading}
              className="w-full"
            >
              {isLoading ? 'LOGGING IN...' : 'LOG IN'}
            </Button>
          </form>

          <div className="w-full py-4">
            <div className="w-full border-t border-parea-black h-px"></div>
          </div>

          <Button
            type="button"
            variant="secondary"
            className="w-full flex items-center justify-center gap-2"
          >
            <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
              <path
                d="M17.64 9.20454C17.64 8.56636 17.5827 7.95272 17.4764 7.36363H9V10.845H13.8436C13.635 11.97 13.0009 12.9232 12.0477 13.5614V15.8195H15.9564C17.4382 14.4186 18.36 12.2727 18.36 9.20454Z"
                fill="#4285F4"
              />
              <path
                d="M9 18C11.43 18 13.467 17.1941 15.9564 15.8195L12.0477 13.5614C11.2418 14.1014 10.2109 14.4204 9 14.4204C6.65454 14.4204 4.67182 12.8373 3.96409 10.71H0.957275V13.0418C2.43818 15.9832 5.48182 18 9 18Z"
                fill="#34A853"
              />
              <path
                d="M3.96409 10.71C3.78409 10.17 3.68182 9.59318 3.68182 9C3.68182 8.40682 3.78409 7.83 3.96409 7.29V4.95818H0.957273C0.347727 6.17318 0 7.54773 0 9C0 10.4523 0.347727 11.8268 0.957273 13.0418L3.96409 10.71Z"
                fill="#FBBC05"
              />
              <path
                d="M9 3.57955C10.3214 3.57955 11.5077 4.03364 12.4405 4.92545L15.0218 2.34409C13.4632 0.891818 11.4259 0 9 0C5.48182 0 2.43818 2.01682 0.957275 4.95818L3.96409 7.29C4.67182 5.16273 6.65454 3.57955 9 3.57955Z"
                fill="#EA4335"
              />
            </svg>
            SIGN UP WITH GOOGLE
          </Button>

          <div className="mt-6 text-center">
            <p className="text-regular text-parea-black">
              Don&apos;t have an account?{' '}
              <Link href="/signup" className="underline">
                Sign Up
              </Link>
            </p>
          </div>
        </div>
      </div>
    </AuthLayout>
  );
}
