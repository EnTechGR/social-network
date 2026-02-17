'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import ProfileWrap from '@/components/ui/ProfileWrap';
import Tabs from '@/components/ui/Tabs';
import PrivateProfileModal from '@/components/ui/PrivateProfileModal';
import { getUserProfile, getAvatarUrl, followUser } from '@/lib/api';

export default function UserProfilePage() {
  const params = useParams<{ id: string }>();
  const id = params?.id as string;
  const [profile, setProfile] = useState<Awaited<ReturnType<typeof getUserProfile>> | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('Posts');
  const [showPrivateModal, setShowPrivateModal] = useState(false);
  const [isSubmittingFollow, setIsSubmittingFollow] = useState(false);
  const [followError, setFollowError] = useState<string | null>(null);
  const [followDone, setFollowDone] = useState(false);

  useEffect(() => {
    if (!id) return;
    let cancelled = false;

    async function load() {
      setIsLoading(true);
      setError(null);
      setFollowError(null);
      setFollowDone(false);
      setIsSubmittingFollow(false);
      try {
        const data = await getUserProfile(id);
        if (cancelled) return;
        setProfile(data);
        if (data.privateProfile) setShowPrivateModal(true);
      } catch (err: unknown) {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : 'Failed to load profile');
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    }

    load();
    return () => { cancelled = true; };
  }, [id]);

  const requestFollow = async () => {
    if (!id || followDone) return;

    setIsSubmittingFollow(true);
    setFollowError(null);

    try {
      await followUser(id);
      setFollowDone(true);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to follow user';
      const normalized = message.toLowerCase();

      if (normalized.includes('already')) {
        setFollowDone(true);
        setFollowError(null);
        return;
      }

      setFollowError(message);
      throw err;
    } finally {
      setIsSubmittingFollow(false);
    }
  };

  if (isLoading) {
    return (
      <main className="min-h-screen bg-parea-white px-16 py-12">
        <div className="max-w-7xl mx-auto">
          <p className="text-regular text-parea-black">Loading profile...</p>
        </div>
      </main>
    );
  }

  if (error) {
    return (
      <main className="min-h-screen bg-parea-white px-16 py-12">
        <div className="max-w-7xl mx-auto">
          <p className="text-regular text-parea-black">{error}</p>
        </div>
      </main>
    );
  }

  if (!profile) {
    return null;
  }

  if (profile.privateProfile) {
    const name = [profile.first_name, profile.last_name].filter(Boolean).join(' ') || profile.nickname || 'User';
    return (
      <main className="min-h-screen bg-parea-white px-16 py-12">
        <div className="max-w-7xl mx-auto">
          <p className="text-regular text-parea-black mb-4">{profile.message}</p>
          <button
            type="button"
            onClick={() => setShowPrivateModal(true)}
            disabled={followDone}
            className="rounded border border-parea-black bg-parea-yellow px-4 py-2 text-small font-medium uppercase text-parea-black hover:opacity-90"
          >
            {followDone ? 'Follow request sent' : 'Send follow request'}
          </button>
          {followError && (
            <p className="text-regular text-parea-black mt-3">{followError}</p>
          )}
        </div>
        <PrivateProfileModal
          isOpen={showPrivateModal}
          onClose={() => setShowPrivateModal(false)}
          userName={name}
          onSendRequest={requestFollow}
          isSubmitting={isSubmittingFollow}
          errorMessage={followError}
        />
      </main>
    );
  }

  const u = profile.user;
  const rawBirth = u.date_of_birth;
  const birthDate = typeof rawBirth === 'string' ? rawBirth.split('T')[0] : '';

  const userForWrap = {
    avatarUrl: getAvatarUrl(u.avatar?.file_path || u.avatar?.thumbnail_path),
    name: `${u.first_name || ''} ${u.last_name || ''}`.trim() || u.nickname || 'User',
    username: u.nickname || (u.email ? u.email.split('@')[0] : ''),
    bio: u.about_me || '',
    email: u.email || '—',
    birthDate,
    isPublic: !u.is_private,
    followersCount: profile.counts.followers,
    followingCount: profile.counts.following,
  };

  const handleFollowersClick = () => {
    // TODO: Open followers modal with real data
  };

  const handleFollowingClick = () => {
    // TODO: Open following modal with real data
  };

  return (
    <main className="min-h-screen bg-parea-white px-16 py-12">
      <div className="max-w-7xl mx-auto">
        <ProfileWrap
          user={userForWrap}
          isSelf={false}
          onFollowersClick={handleFollowersClick}
          onFollowingClick={handleFollowingClick}
        />

        {!profile.is_own_profile && (
          <div className="mt-4 flex flex-col items-start gap-2">
            <button
              type="button"
              onClick={() => {
                void requestFollow().catch(() => {});
              }}
              disabled={isSubmittingFollow || followDone}
              className="rounded border border-parea-black bg-parea-yellow px-4 py-2 text-small font-medium uppercase text-parea-black hover:opacity-90 disabled:opacity-60 disabled:cursor-not-allowed"
            >
              {followDone ? 'Following' : isSubmittingFollow ? 'Following...' : 'Follow'}
            </button>
            {followError && (
              <p className="text-regular text-parea-black">{followError}</p>
            )}
          </div>
        )}

        <div className="mt-8">
          <Tabs
            tabs={['Posts', 'Events', 'Reactions', 'Groups']}
            defaultTab="Posts"
            onTabChange={(tab) => setActiveTab(tab)}
          />
          <div className="mt-6">
            {activeTab === 'Posts' && <p>Posts content...</p>}
            {activeTab === 'Events' && <p>Events content...</p>}
            {activeTab === 'Reactions' && <p>Reactions content...</p>}
            {activeTab === 'Groups' && <p>Groups content...</p>}
          </div>
        </div>
      </div>
    </main>
  );
}
