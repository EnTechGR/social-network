'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import ProfileWrap from '@/components/ui/ProfileWrap';
import Tabs from '@/components/ui/Tabs';
import PrivateProfileModal from '@/components/ui/PrivateProfileModal';
import Card from '@/components/ui/Card';
import FollowersModal from '@/components/ui/FollowersModal';
import {
  getUserProfile,
  getAvatarUrl,
  followUser,
  getProfile,
  unfollowUser,
  getPostById,
  getPostImageUrl,
} from '@/lib/api';

function formatPostDate(isoDate: string): string {
  if (!isoDate) return '';
  try {
    const d = new Date(isoDate);
    return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
  } catch {
    return isoDate;
  }
}

function formatRelationName(entry: any): string {
  const fullName = [entry?.first_name, entry?.last_name].filter(Boolean).join(' ').trim();
  return fullName || entry?.nickname || entry?.email || entry?.user_id || 'Unknown user';
}

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
  const [followRequested, setFollowRequested] = useState(false);
  const [isFollowing, setIsFollowing] = useState(false);
  const [showFollowersList, setShowFollowersList] = useState(false);
  const [showFollowingList, setShowFollowingList] = useState(false);
  const [showFollowersModal, setShowFollowersModal] = useState(false);
  const [postImagesById, setPostImagesById] = useState<Record<string, string | undefined>>({});
  const [showFollowingModal, setShowFollowingModal] = useState(false);

  useEffect(() => {
    if (!id) return;
    let cancelled = false;

    async function load() {
      setIsLoading(true);
      setError(null);
      setFollowError(null);
      setFollowRequested(false);
      setIsFollowing(false);
      setIsSubmittingFollow(false);
      try {
        const [data, me] = await Promise.all([getUserProfile(id), getProfile()]);
        if (cancelled) return;
        setProfile(data);
        if (data.privateProfile) setShowPrivateModal(true);
        if (!data.privateProfile && !data.is_own_profile && me?.id) {
          const followerIds = Array.isArray(data.followers)
            ? data.followers.map((f) => (f as any)?.user_id).filter(Boolean)
            : [];
          setIsFollowing(followerIds.includes(me.id));
        }
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

  // Enrich this user's posts with the same image URLs used by the main feed,
  // by fetching each post's full detail once profile.posts is available.
  useEffect(() => {
    if (!profile || profile.privateProfile) return;
    const posts = Array.isArray(profile.posts) ? profile.posts : [];
    if (posts.length === 0) {
      setPostImagesById({});
      return;
    }

    let cancelled = false;

    (async () => {
      try {
        const uniqueIds = Array.from(
          new Set(
            posts
              .map((p: any) => p.id || p.post_id)
              .filter((id: unknown): id is string => typeof id === 'string' && id.length > 0),
          ),
        );

        if (uniqueIds.length === 0) {
          if (!cancelled) setPostImagesById({});
          return;
        }

        const entries = await Promise.all(
          uniqueIds.map(async (postId) => {
            try {
              const detail = await getPostById(postId);
              const firstImage = detail?.images?.[0];
              const rawUrl = firstImage?.thumbnail_url || firstImage?.url || undefined;
              const imageUrl = getPostImageUrl(rawUrl);
              return [postId, imageUrl] as const;
            } catch {
              return [postId, undefined] as const;
            }
          }),
        );

        if (cancelled) return;

        const map: Record<string, string | undefined> = {};
        for (const [postId, imageUrl] of entries) {
          map[postId] = imageUrl;
        }
        setPostImagesById(map);
      } catch {
        if (!cancelled) setPostImagesById({});
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [profile]);

  const requestFollow = async () => {
    if (!id || followRequested || isFollowing) return;

    setIsSubmittingFollow(true);
    setFollowError(null);

    try {
      const relationship = await followUser(id);
      if (relationship.status === 'accepted') {
        setIsFollowing(true);
        setFollowRequested(false);
      } else {
        setFollowRequested(true);
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to follow user';
      const normalized = message.toLowerCase();

      if (normalized.includes('already')) {
        if (profile && !profile.privateProfile) {
          setIsFollowing(true);
        } else {
          setFollowRequested(true);
        }
        setFollowError(null);
        return;
      }

      setFollowError(message);
      throw err;
    } finally {
      setIsSubmittingFollow(false);
    }
  };

  const handleUnfollow = async () => {
    if (!id || !isFollowing) return;

    setIsSubmittingFollow(true);
    setFollowError(null);
    try {
      await unfollowUser(id);
      setIsFollowing(false);
      setFollowRequested(false);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to unfollow user';
      setFollowError(message);
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
            disabled={followRequested}
            className="rounded border border-parea-black bg-parea-yellow px-4 py-2 text-small font-medium uppercase text-parea-black hover:opacity-90"
          >
            {followRequested ? 'Follow request sent' : 'Send follow request'}
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
    gender: u.gender || '',
    isPublic: !u.is_private,
    followersCount: profile.counts.followers,
    followingCount: profile.counts.following,
  };

  const handleFollowersClick = () => {
    setShowFollowersModal(true);
  };

  const handleFollowingClick = () => {
    setShowFollowingModal(true);
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
                if (isFollowing) {
                  void handleUnfollow();
                } else {
                  void requestFollow().catch(() => {});
                }
              }}
              disabled={isSubmittingFollow || followRequested}
              className="rounded border border-parea-black bg-parea-yellow px-4 py-2 text-small font-medium uppercase text-parea-black hover:opacity-90 disabled:opacity-60 disabled:cursor-not-allowed"
            >
              {isFollowing
                ? (isSubmittingFollow ? 'Unfollowing...' : 'Unfollow')
                : (followRequested ? 'Follow request sent' : (isSubmittingFollow ? 'Following...' : 'Follow'))}
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
            {activeTab === 'Posts' && (
              <>
                {!Array.isArray(profile.posts) || profile.posts.length === 0 ? (
                  <p className="text-regular text-parea-black">No posts yet.</p>
                ) : (
                  <div className="flex flex-col items-start gap-0">
                    {profile.posts.map((post: any, index: number) => {
                      const postId = post.id || post.post_id;
                      const enrichedImage = postId ? postImagesById[postId] : undefined;
                      const rawImagePath = enrichedImage || post.image_url || post.thumbnail_url;
                      const imageSrc = getPostImageUrl(rawImagePath);
                      return (
                        <Card
                          key={postId || `${post.title}-${index}`}
                          imageType="post"
                          imageSrc={imageSrc}
                          avatarSrc={getAvatarUrl(u.avatar?.thumbnail_path || u.avatar?.file_path)}
                          avatarAlt={u.nickname || 'Author'}
                          userName={(u.nickname || 'User').toUpperCase()}
                          userDate={formatPostDate(post.created_at)}
                          title={post.title}
                          content={post.content}
                          href={postId ? `/post/${postId}` : undefined}
                          imagePriority={index === 0}
                          likeCount={post.like_count || 0}
                          commentCount={post.comment_count || 0}
                        />
                      );
                    })}
                  </div>
                )}
              </>
            )}
            {activeTab === 'Events' && <p>Events content...</p>}
            {activeTab === 'Reactions' && <p>Reactions content...</p>}
            {activeTab === 'Groups' && <p>Groups content...</p>}
          </div>
        </div>

        <FollowersModal
          isOpen={showFollowersModal}
          onClose={() => setShowFollowersModal(false)}
          heading="Followers"
          users={Array.isArray(profile.followers) ? profile.followers : []}
        />

        <FollowersModal
          isOpen={showFollowingModal}
          onClose={() => setShowFollowingModal(false)}
          heading="Following"
          users={Array.isArray(profile.following) ? profile.following : []}
        />
      </div>
    </main>
  );
}
