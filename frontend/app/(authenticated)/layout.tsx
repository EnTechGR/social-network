'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { verifySession } from '@/lib/auth';
import Sidebar from '@/components/ui/Sidebar';
import Navbar from '@/components/ui/Navbar';

export default function AuthenticatedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const skipAuth = true; // Set to false to enable auth
  const [isAuthenticated, setIsAuthenticated] = useState(skipAuth);

  useEffect(() => {
    if (skipAuth) return;

    const checkAuth = async () => {
      const session = await verifySession();
      if (!session) {
        router.push('/login');
      } else {
        setIsAuthenticated(true);
      }
    };
    checkAuth();
  }, [router, skipAuth]);

  if (!isAuthenticated) {
    return (
      <div className="fixed inset-0 z-10000 bg-parea-white flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-parea-black border-r-transparent align-[-0.125em] motion-reduce:animate-[spin_1.5s_linear_infinite]" />
          <p className="mt-4 text-parea-black">Loading...</p>
        </div>
      </div>
    );
  }

  return (
    <>
      <Sidebar />
      <div className="ml-18">
        <Navbar />
        <main>{children}</main>
      </div>
    </>
  );
}
