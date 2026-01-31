'use client';

import { useParams } from 'next/navigation';

export default function UserProfilePage() {
  const { id } = useParams<{ id: string }>();

  return (
    <main className="min-h-screen bg-parea-white px-16 py-12">
      <div className="max-w-7xl mx-auto">
        <p className="text-regular text-parea-black">User profile: {id}</p>
      </div>
    </main>
  );
}
