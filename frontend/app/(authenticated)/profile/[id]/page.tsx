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
  getChatConversation,
  sendChatMessage,
  type DirectChatMessage,
} from '@/lib/api';
import Button from '@/components/ui/Button';
import ChatModal from '@/components/ui/ChatModal';

function formatPostDate(isoDate: string): string {
  if (!isoDate) return '';
  try {
    const d = new Date(isoDate);
    return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
  } catch {
    return isoDate;
  }
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
  const [showFollowersModal, setShowFollowersModal] = useState(false);
  const [postMetaById, setPostMetaById] = useState<Record<string, { imageUrl?: string; commentCount?: number }>>({});
  const [showFollowingModal, setShowFollowingModal] = useState(false);
  const [chatModalOpen, setChatModalOpen] = useState(false);
  const [chatMessages, setChatMessages] = useState<DirectChatMessage[]>([]);
  const [currentUserId, setCurrentUserId] = useState('');
  const [isChatSending, setIsChatSending] = useState(false);
  const [selfAvatarUrl, setSelfAvatarUrl] = useState('/user-avatar-default.png');

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
        if (me?.id) setCurrentUserId(typeof me.id === 'string' ? me.id : '');
        setSelfAvatarUrl(getAvatarUrl(me?.avatar?.thumbnail_path || me?.avatar?.file_path));
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
      setPostMetaById({});
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
          if (!cancelled) setPostMetaById({});
          return;
        }

        const entries = await Promise.all(
          uniqueIds.map(async (postId) => {
            try {
              const detail = await getPostById(postId);
              const firstImage = detail?.images?.[0];
              const rawUrl = firstImage?.thumbnail_url || firstImage?.url || undefined;
              const imageUrl = getPostImageUrl(rawUrl);
              const commentCount = typeof detail?.comment_count === 'number' ? detail.comment_count : undefined;
              return [postId, { imageUrl, commentCount }] as const;
            } catch {
              return [postId, {}] as const;
            }
          }),
        );

        if (cancelled) return;

        const map: Record<string, { imageUrl?: string; commentCount?: number }> = {};
        for (const [postId, meta] of entries) {
          map[postId] = meta;
        }
        setPostMetaById(map);
      } catch {
        if (!cancelled) setPostMetaById({});
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

  const handleOpenChat = async () => {
    setChatModalOpen(true);
    try {
      const messages = await getChatConversation(id, 100, 0);
      setChatMessages(messages);
    } catch {
      setChatMessages([]);
    }
  };

  const handleSendChatMessage = async (text: string) => {
    if (isChatSending) return;
    setIsChatSending(true);
    try {
      const message = await sendChatMessage(id, text);
      setChatMessages((prev) => [...prev, message]);
    } catch {
      // message failed silently
    } finally {
      setIsChatSending(false);
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
    const privateProfileAvatarPath =
      (profile as any)?.avatar?.thumbnail_path ||
      (profile as any)?.avatar?.file_path ||
      (profile as any)?.user?.avatar?.thumbnail_path ||
      (profile as any)?.user?.avatar?.file_path;
    const privateProfileAvatarSrc = privateProfileAvatarPath ? getAvatarUrl(privateProfileAvatarPath) : undefined;
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
          avatarSrc={privateProfileAvatarSrc}
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

        <div className="mt-8 max-w-312 w-full flex items-center justify-between">
          <Tabs
            tabs={['Posts', 'Events', 'Groups']}
            defaultTab="Posts"
            onTabChange={(tab) => setActiveTab(tab)}
          />
          {!profile.is_own_profile && (
            <div className="flex items-center gap-3">
              <Button variant="primary" size="md" onClick={() => { void handleOpenChat(); }}>
                Chat
              </Button>
              <Button
                variant="secondary"
                size="md"
                disabled={isSubmittingFollow}
                onClick={() => {
                  if (isFollowing) void handleUnfollow();
                  else void requestFollow().catch(() => { });
                }}
              >
                <span style={{ display: 'block', textAlign: 'center', width: '5rem' }}>
                  {isFollowing
                    ? (isSubmittingFollow ? 'Unfollowing...' : 'Unfollow')
                    : (followRequested ? 'Requested' : (isSubmittingFollow ? 'Following...' : 'Follow'))}
                </span>
              </Button>
            </div>
          )}
        </div>

        <div className="mt-6 max-w-312 w-full">
          {activeTab === 'Posts' && (
            <>
              {!Array.isArray(profile.posts) || profile.posts.length === 0 ? (
                <p className="text-regular text-parea-black">No posts yet.</p>
              ) : (
                <div className="flex flex-col items-start gap-0">
                  {profile.posts.map((post: any, index: number) => {
                    const postId = post.id || post.post_id;
                    const meta = postId ? (postMetaById[postId] ?? {}) : {};
                    const rawImagePath = meta.imageUrl || post.image_url || post.thumbnail_url;
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
                        commentCount={meta.commentCount ?? 0}
                      />
                    );
                  })}
                </div>
              )}
            </>
          )}
          {activeTab === 'Events' && <p className="text-regular text-parea-black">No events yet.</p>}

          {activeTab === 'Groups' && <p className="text-regular text-parea-black">No groups yet.</p>}
        </div>

        <ChatModal
          isOpen={chatModalOpen}
          onClose={() => setChatModalOpen(false)}
          userName={userForWrap.username}
          userAvatar={userForWrap.avatarUrl}
          selfAvatar={selfAvatarUrl}
          controlledMessages={chatMessages.map((m) => ({
            id: m.message_id,
            text: m.content,
            sender: m.sender_id === currentUserId ? 'self' : 'other',
          }))}
          onSendMessage={(text) => { void handleSendChatMessage(text); }}
        />

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
