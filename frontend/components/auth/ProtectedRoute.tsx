/**
 * components/auth/ProtectedRoute.tsx
 *
 * A wrapper component that protects pages from unauthorized access.
 * Redirects to login if user is not authenticated.
 */

'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { verifySession } from '@/lib/auth';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

export default function ProtectedRoute({ children }: ProtectedRouteProps) {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Verify authentication with the API
    const checkAuth = async () => {
      const session = await verifySession();
      
      if (!session) {
        // No valid session - redirect to login
        router.push('/login');
      } else {
        // Valid session - allow access
        setIsLoading(false);
      }
    };

    checkAuth();
  }, [router]);

  // Show full-screen loading overlay while checking authentication
  // This prevents unauthorized users from seeing any content
  if (isLoading) {
    return (
      <div className="fixed inset-0 z-[10000] bg-parea-white flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-parea-black border-r-transparent align-[-0.125em] motion-reduce:animate-[spin_1.5s_linear_infinite]"></div>
          <p className="mt-4 text-parea-black">Loading...</p>
        </div>
      </div>
    );
  }

  // Render children if authenticated
  return <>{children}</>;
}
