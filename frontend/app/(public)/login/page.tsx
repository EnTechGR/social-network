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
