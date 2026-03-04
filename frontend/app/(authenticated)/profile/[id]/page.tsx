'use client';

import { use, useState, useEffect } from 'react';
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
  getPostsByUserId,
  getPostImageUrl,
  getChatConversation,
  sendChatMessage,
  getAllGroups,
  getGroupMembers,
  type DirectChatMessage,
} from '@/lib/api';
import Button from '@/components/ui/Button';
import ChatModal from '@/components/ui/ChatModal';

const PENDING_FOLLOW_STORAGE_KEY = 'follow_requests_pending';

function getPendingFollowIds(): string[] {
  if (typeof window === 'undefined') return [];
  try {
    const raw = localStorage.getItem(PENDING_FOLLOW_STORAGE_KEY);
    const parsed = raw ? JSON.parse(raw) : [];
    return Array.isArray(parsed) ? parsed.filter((x): x is string => typeof x === 'string') : [];
  } catch {
    return [];
  }
}

function addPendingFollow(userId: string) {
  const ids = getPendingFollowIds();
  if (ids.includes(userId)) return;
  ids.push(userId);
  localStorage.setItem(PENDING_FOLLOW_STORAGE_KEY, JSON.stringify(ids));
}

function removePendingFollow(userId: string) {
  const ids = getPendingFollowIds().filter((x) => x !== userId);
  localStorage.setItem(PENDING_FOLLOW_STORAGE_KEY, JSON.stringify(ids));
}

function formatPostDate(isoDate: string): string {
  if (!isoDate) return '';
  try {
    const d = new Date(isoDate);
    return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
  } catch {
    return isoDate;
  }
}

export default function UserProfilePage({
  params,
}: {
  params: Promise<{ id?: string }>;
}) {
  const resolvedParams = use(params);
  const id = resolvedParams?.id as string;
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
  const [profilePosts, setProfilePosts] = useState<any[]>([]);
  const [refetching, setRefetching] = useState(false);
  const [userGroups, setUserGroups] = useState<any[]>([]);
  const [groupsLoading, setGroupsLoading] = useState(false);
  const [groupsError, setGroupsError] = useState<string | null>(null);

  const refetchProfile = async () => {
    if (!id) return;
    setRefetching(true);
    try {
      const [data, me] = await Promise.all([getUserProfile(id), getProfile()]);
      setProfile(data);
      if (me?.id) setCurrentUserId(typeof me.id === 'string' ? me.id : '');
      setSelfAvatarUrl(getAvatarUrl(me?.avatar?.thumbnail_path || me?.avatar?.file_path));
      if (data.privateProfile) {
        setShowPrivateModal(true);
        setFollowRequested(getPendingFollowIds().includes(id));
      } else {
        removePendingFollow(id);
        setFollowRequested(false);
        if (!data.is_own_profile && me?.id) {
          const followerIds = Array.isArray(data.followers)
            ? data.followers.map((f) => (f as any)?.user_id).filter(Boolean)
            : [];
          setIsFollowing(followerIds.includes(me.id));
        }
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load profile');
    } finally {
      setRefetching(false);
    }
  };

  useEffect(() => {
    if (!id) return;
    let cancelled = false;

    async function load() {
      setIsLoading(true);
      setError(null);
      setFollowError(null);
      setIsSubmittingFollow(false);
      try {
        const [data, me] = await Promise.all([getUserProfile(id), getProfile()]);
        if (cancelled) return;
        setProfile(data);
        if (me?.id) setCurrentUserId(typeof me.id === 'string' ? me.id : '');
        setSelfAvatarUrl(getAvatarUrl(me?.avatar?.thumbnail_path || me?.avatar?.file_path));
        if (data.privateProfile) {
          setShowPrivateModal(true);
          setFollowRequested(getPendingFollowIds().includes(id));
          setIsFollowing(false);
        } else {
          removePendingFollow(id);
          setFollowRequested(false);
          if (!data.is_own_profile && me?.id) {
            const followerIds = Array.isArray(data.followers)
              ? data.followers.map((f) => (f as any)?.user_id).filter(Boolean)
              : [];
            setIsFollowing(followerIds.includes(me.id));
          }
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

  useEffect(() => {
    if (!id || !profile?.privateProfile) return;
    const onFocus = () => refetchProfile();
    window.addEventListener('focus', onFocus);
    return () => window.removeEventListener('focus', onFocus);
  }, [id, profile?.privateProfile]);

  useEffect(() => {
    if (!id || !profile || profile.privateProfile) {
      setProfilePosts([]);
      return;
    }
    let cancelled = false;
    getPostsByUserId(id)
      .then((list) => {
        if (!cancelled) setProfilePosts(Array.isArray(list) ? list : []);
      })
      .catch(() => {
        if (!cancelled) setProfilePosts([]);
      });
    return () => { cancelled = true; };
  }, [id, profile]);

  useEffect(() => {
    if (!id || !profile || profile.privateProfile || activeTab !== 'Groups') {
      return;
    }

    let cancelled = false;
    setGroupsLoading(true);
    setGroupsError(null);

    (async () => {
      try {
        const allGroups = await getAllGroups();
        const groups = Array.isArray(allGroups) ? allGroups : [];

        const membershipChecks = await Promise.all(
          groups.map(async (group: any) => {
            const groupId = group?.id;
            if (typeof groupId !== 'string' || groupId.length === 0) {
              return null;
            }

            try {
              const members = await getGroupMembers(groupId);
              const list = Array.isArray(members) ? members : [];
              const isMember = list.some((member: any) => {
                const memberId = member?.user_id ?? member?.id;
                return memberId === id;
              });
              return isMember ? group : null;
            } catch {
              return null;
            }
          }),
        );

        if (cancelled) return;
        setUserGroups(membershipChecks.filter(Boolean));
      } catch (err: unknown) {
        if (cancelled) return;
        setUserGroups([]);
        setGroupsError(err instanceof Error ? err.message : 'Failed to load groups');
      } finally {
        if (!cancelled) setGroupsLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [activeTab, id, profile]);

  // Enrich this user's posts with the same image URLs used by the main feed,
  // by fetching each post's full detail once profilePosts is available.
  useEffect(() => {
    if (!profile || profile.privateProfile) return;
    const posts = profilePosts;
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
            const postFromList = posts.find((p: any) => (p?.id || p?.post_id) === postId);
            const listCommentCount =
              typeof postFromList?.comment_count === 'number'
                ? postFromList.comment_count
                : Array.isArray(postFromList?.comments)
                  ? postFromList.comments.length
                  : undefined;

            try {
              const detail = await getPostById(postId);
              const firstImage = detail?.images?.[0];
              const rawUrl = firstImage?.thumbnail_url || firstImage?.url || undefined;
              const imageUrl = getPostImageUrl(rawUrl) || getPostImageUrl(postFromList?.thumbnail_url || postFromList?.image_url);
              const commentCount =
                typeof detail?.comment_count === 'number'
                  ? detail.comment_count
                  : listCommentCount;
              return [postId, { imageUrl, commentCount }] as const;
            } catch {
              const imageUrl = getPostImageUrl(postFromList?.thumbnail_url || postFromList?.image_url);
              return [postId, { imageUrl, commentCount: listCommentCount }] as const;
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
  }, [profile, profilePosts]);

  const requestFollow = async () => {
    if (!id || followRequested || isFollowing) return;

    setIsSubmittingFollow(true);
    setFollowError(null);

    try {
      const relationship = await followUser(id);
      if (relationship.status === 'accepted') {
        setIsFollowing(true);
        setFollowRequested(false);
        removePendingFollow(id);
      } else {
        setFollowRequested(true);
        addPendingFollow(id);
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to follow user';
      const normalized = message.toLowerCase();

      if (normalized.includes('already')) {
        if (profile && !profile.privateProfile) {
          setIsFollowing(true);
          removePendingFollow(id);
        } else {
          setFollowRequested(true);
          addPendingFollow(id);
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
      removePendingFollow(id);
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
    const privateProfileAvatarRaw =
      (profile as any)?.avatar?.thumbnail_path ||
      (profile as any)?.avatar?.file_path ||
      (profile as any)?.avatar_thumbnail_path ||
      (profile as any)?.avatar_path ||
      (profile as any)?.avatar_thumb_url ||
      (profile as any)?.avatar_url ||
      (profile as any)?.user?.avatar?.thumbnail_path ||
      (profile as any)?.user?.avatar?.file_path;
    const privateProfileAvatarSrc = privateProfileAvatarRaw
      ? (/^https?:\/\//i.test(privateProfileAvatarRaw) ? privateProfileAvatarRaw : getAvatarUrl(privateProfileAvatarRaw))
      : undefined;
    return (
      <main className="min-h-screen bg-parea-white px-16 py-12">
        <div className="max-w-7xl mx-auto">
          <p className="text-regular text-parea-black mb-4">{profile.message}</p>
          <div className="flex flex-wrap items-center gap-2">
            <button
              type="button"
              onClick={() => setShowPrivateModal(true)}
              disabled={followRequested}
              className="rounded border border-parea-black bg-parea-yellow px-4 py-2 text-small font-medium uppercase text-parea-black hover:opacity-90"
            >
              {followRequested ? 'Follow request sent' : 'Send follow request'}
            </button>
            {followRequested && (
              <button
                type="button"
                onClick={() => refetchProfile()}
                disabled={refetching}
                className="rounded border border-parea-black bg-parea-white px-4 py-2 text-small font-medium uppercase text-parea-black hover:opacity-90 disabled:opacity-60"
              >
                {refetching ? 'Checking...' : 'Check again'}
              </button>
            )}
          </div>
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
            tabs={['Posts', 'Groups']}
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
              {profilePosts.length === 0 ? (
                <p className="text-regular text-parea-black">No posts yet.</p>
              ) : (
                <div className="flex flex-col items-start gap-0">
                  {profilePosts.map((post: any, index: number) => {
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
          {activeTab === 'Groups' && (
            <>
              {groupsLoading && <p className="text-regular text-parea-black">Loading groups...</p>}
              {!groupsLoading && groupsError && <p className="text-regular text-parea-black">{groupsError}</p>}
              {!groupsLoading && !groupsError && userGroups.length === 0 && (
                <p className="text-regular text-parea-black">No groups yet.</p>
              )}
              {!groupsLoading && !groupsError && userGroups.length > 0 && (
                <div className="flex flex-col gap-4">
                  {userGroups.map((g) => (
                    <a
                      key={g.id}
                      href={`/group/${g.id}`}
                      className="border border-parea-black p-6 bg-white cursor-pointer hover:shadow-[4px_4px_0_0_#000] transition-shadow no-underline"
                    >
                      <h3 className="text-2xl font-bold text-parea-black mb-2">
                        {g.title ?? g.name ?? 'Group'}
                      </h3>
                      {g.description && (
                        <p className="text-regular text-parea-black/70 mb-3">
                          {g.description}
                        </p>
                      )}
                      <div className="flex gap-4 text-sm text-parea-black/60">
                        <span>{g.member_count || 0} members</span>
                        <span>•</span>
                        <span>Created by {g.owner_nickname || 'Unknown'}</span>
                      </div>
                    </a>
                  ))}
                </div>
              )}
            </>
          )}
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
