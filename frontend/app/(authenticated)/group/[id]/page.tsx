'use client';

import { useRouter } from 'next/navigation';
import { use, useState, useEffect } from 'react';
import GroupWrap from '@/components/ui/GroupWrap';
import Tabs from '@/components/ui/Tabs';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import CreatePostModal from '@/components/ui/CreatePostModal';
import CreateEventModal from '@/components/ui/CreateEventModal';
import FollowersModal, { type FollowerUser } from '@/components/ui/FollowersModal';
import {
  getGroupById,
  getGroupMembers,
  getGroupPosts,
  getGroupEvents,
  getProfile,
  getUserProfile,
  inviteToGroup,
  leaveGroup,
  requestToJoinGroup,
  getPendingGroupRequests,
  approveGroupRequest,
  denyGroupRequest,
  getAvatarUrl,
  getPostImageUrl,
  getPostById,
} from '@/lib/api';

function formatDate(iso: string): string {
  if (!iso) return '';
  try {
    const d = new Date(iso);
    return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
  } catch {
    return iso;
  }
}



export default function GroupDetailPage({
  params,
}: {
  params: Promise<{ id?: string }>;
}) {
  const resolvedParams = use(params);
  const router = useRouter();
  const groupId = typeof resolvedParams?.id === 'string' ? resolvedParams.id : '';
  const [group, setGroup] = useState<any | null>(null);
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [isMember, setIsMember] = useState(false);
  const [isOwner, setIsOwner] = useState(false);
  const [posts, setPosts] = useState<any[]>([]);
  const [events, setEvents] = useState<any[]>([]);
  const [members, setMembers] = useState<any[]>([]);
  const [pendingRequests, setPendingRequests] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('Posts');
  const [postsLoading, setPostsLoading] = useState(false);
  const [eventsLoading, setEventsLoading] = useState(false);
  const [createPostOpen, setCreatePostOpen] = useState(false);
  const [createEventOpen, setCreateEventOpen] = useState(false);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [showMembersModal, setShowMembersModal] = useState(false);
  const [followers, setFollowers] = useState<FollowerUser[]>([]);
  const [inviteMessage, setInviteMessage] = useState<string | null>(null);
  const [inviteLoading, setInviteLoading] = useState<string | null>(null);
  const [joinLoading, setJoinLoading] = useState(false);
  const [joinPending, setJoinPending] = useState(false);
  const [requestActionId, setRequestActionId] = useState<string | null>(null);
  const [postMetaById, setPostMetaById] = useState<Record<string, { imageUrl?: string; avatarUrl?: string; commentCount?: number }>>({});
  const [eventCreatorAvatarById, setEventCreatorAvatarById] = useState<Record<string, string>>({});

  useEffect(() => {
    if (!groupId) {
      setError('Invalid group');
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    setError(null);
    Promise.all([getGroupById(groupId), getProfile()])
      .then(([g, profile]) => {
        if (cancelled) return;
        setGroup(g);
        const pid = (profile as any)?.id ?? (profile as any)?.user_id ?? (profile as any)?.user?.id;
        setCurrentUserId(pid ?? null);
        setIsOwner(Boolean(g?.owner_id && pid && g.owner_id === pid));
      })
      .catch((err: any) => {
        if (cancelled) return;
        setError(err?.message ?? 'Failed to load group');
        setGroup(null);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, [groupId]);

  useEffect(() => {
    if (!groupId || !group) return;
    let cancelled = false;
    getGroupMembers(groupId)
      .then((list) => {
        if (cancelled) return;
        setIsMember(true);
        setMembers(Array.isArray(list) ? list : []);
      })
      .catch(() => {
        if (cancelled) return;
        setIsMember(false);
        setMembers([]);
      });
    return () => { cancelled = true; };
  }, [groupId, group]);

  useEffect(() => {
    if (!group) return;
    setIsOwner(Boolean(group.owner_id && currentUserId && group.owner_id === currentUserId));
  }, [group, currentUserId]);

  useEffect(() => {
    if (!groupId || !isOwner) {
      setPendingRequests([]);
      return;
    }
    let cancelled = false;
    getPendingGroupRequests(groupId)
      .then((requests) => {
        if (!cancelled) setPendingRequests(Array.isArray(requests) ? requests : []);
      })
      .catch(() => {
        if (!cancelled) setPendingRequests([]);
      });
    return () => { cancelled = true; };
  }, [groupId, isOwner]);

  useEffect(() => {
    if (!groupId || !isMember) return;
    let cancelled = false;
    setPostsLoading(true);
    getGroupPosts(groupId)
      .then((list) => {
        if (!cancelled) setPosts(Array.isArray(list) ? list : []);
      })
      .catch(() => {
        if (!cancelled) setPosts([]);
      })
      .finally(() => {
        if (!cancelled) setPostsLoading(false);
      });
    return () => { cancelled = true; };
  }, [groupId, isMember]);

  // Enrich group posts with the same image and avatar data used by the main feed
  // by loading each post's full detail once the list is available.
  useEffect(() => {
    if (!isMember || posts.length === 0) return;
    let cancelled = false;

    (async () => {
      try {
        const uniqueIds = Array.from(new Set((posts ?? []).map((p: any) => p.id).filter(Boolean)));
        if (uniqueIds.length === 0) {
          if (!cancelled) setPostMetaById({});
          return;
        }

        const entries = await Promise.all(
          uniqueIds.map(async (postId) => {
            try {
              const detail = await getPostById(postId);

              // Feed detail already returns fully-qualified avatar and image URLs,
              // so we can use them directly without further transformation.
              const avatarUrl =
                detail?.author_avatar_thumb_url || detail?.author_avatar_url || undefined;

              const firstImage = detail?.images?.[0];
              const imageUrl = firstImage?.thumbnail_url || firstImage?.url || undefined;

              const commentCount =
                typeof detail?.comment_count === 'number'
                  ? detail.comment_count
                  : undefined;

              return [postId, { avatarUrl, imageUrl, commentCount }] as const;
            } catch {
              return [postId, { avatarUrl: undefined, imageUrl: undefined, commentCount: undefined }] as const;
            }
          }),
        );

        if (cancelled) return;

        const map: Record<string, { imageUrl?: string; avatarUrl?: string; commentCount?: number }> = {};
        for (const [id, meta] of entries) {
          map[id] = meta;
        }
        setPostMetaById(map);
      } catch {
        if (!cancelled) setPostMetaById({});
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [isMember, posts]);

  useEffect(() => {
    if (!groupId || !isMember) return;
    let cancelled = false;
    setEventsLoading(true);
    getGroupEvents(groupId)
      .then((list) => {
        if (!cancelled) setEvents(Array.isArray(list) ? list : []);
      })
      .catch(() => {
        if (!cancelled) setEvents([]);
      })
      .finally(() => {
        if (!cancelled) setEventsLoading(false);
      });
    return () => { cancelled = true; };
  }, [groupId, isMember]);

  useEffect(() => {
    if (!Array.isArray(events) || events.length === 0) {
      setEventCreatorAvatarById({});
      return;
    }

    let cancelled = false;

    (async () => {
      const creatorIDs = Array.from(
        new Set(
          events
            .map((event: any) => event?.creator_id)
            .filter((id: unknown): id is string => typeof id === 'string' && id.length > 0),
        ),
      );

      if (creatorIDs.length === 0) {
        if (!cancelled) setEventCreatorAvatarById({});
        return;
      }

      const entries = await Promise.all(
        creatorIDs.map(async (creatorID) => {
          try {
            const profile = await getUserProfile(creatorID);
            const profileAny = profile as any;
            const avatarPath =
              profileAny?.user?.avatar?.thumbnail_path ||
              profileAny?.user?.avatar?.file_path ||
              profileAny?.avatar?.thumbnail_path ||
              profileAny?.avatar?.file_path;
            const avatarUrl = avatarPath ? getAvatarUrl(avatarPath) : '/user-avatar-default.png';
            return [creatorID, avatarUrl] as const;
          } catch {
            return [creatorID, '/user-avatar-default.png'] as const;
          }
        }),
      );

      if (cancelled) return;

      const map: Record<string, string> = {};
      for (const [creatorID, avatarUrl] of entries) {
        map[creatorID] = avatarUrl;
      }
      setEventCreatorAvatarById(map);
    })();

    return () => {
      cancelled = true;
    };
  }, [events]);


  const [membersModalUsers, setMembersModalUsers] = useState<FollowerUser[]>([]);

  useEffect(() => {
    if (!showMembersModal || !groupId) return;
    let cancelled = false;
    (async () => {
      try {
        const list = await getGroupMembers(groupId);
        const rawMembers = Array.isArray(list) ? list : [];
        const membersWithAvatar = await Promise.all(
          rawMembers.map(async (m: any) => {
            const baseUser: FollowerUser = {
              user_id: m.user_id,
              nickname: m.nickname,
              first_name: m.first_name,
              last_name: m.last_name,
              email: m.email,
              avatar: m.avatar,
              avatar_path: m.avatar_path,
              avatar_thumbnail_path: m.avatar_thumbnail_path,
              avatar_url: m.avatar_url,
              avatar_thumb_url: m.avatar_thumb_url,
            };

            const hasAvatarData = Boolean(
              m?.avatar?.thumbnail_path ||
              m?.avatar?.file_path ||
              m?.avatar_thumbnail_path ||
              m?.avatar_path ||
              m?.avatar_thumb_url ||
              m?.avatar_url,
            );

            if (hasAvatarData || !m?.user_id) {
              return baseUser;
            }

            try {
              const profile = await getUserProfile(m.user_id);
              const p = profile as any;
              return {
                ...baseUser,
                avatar: baseUser.avatar ?? p?.user?.avatar ?? p?.avatar,
                avatar_path:
                  baseUser.avatar_path ??
                  p?.avatar_path ??
                  p?.user?.avatar?.file_path ??
                  p?.avatar?.file_path,
                avatar_thumbnail_path:
                  baseUser.avatar_thumbnail_path ??
                  p?.avatar_thumbnail_path ??
                  p?.user?.avatar?.thumbnail_path ??
                  p?.avatar?.thumbnail_path,
                avatar_url: baseUser.avatar_url ?? p?.avatar_url,
                avatar_thumb_url: baseUser.avatar_thumb_url ?? p?.avatar_thumb_url,
              } as FollowerUser;
            } catch {
              return baseUser;
            }
          }),
        );

        if (!cancelled) {
          setMembersModalUsers(membersWithAvatar);
        }
      } catch {
        if (!cancelled) setMembersModalUsers([]);
      }
    })();
    return () => { cancelled = true; };
  }, [showMembersModal, groupId]);

  const handleLeaveGroup = async () => {
    if (!groupId) return;
    try {
      await leaveGroup(groupId);
      setIsMember(false);
      setMembers([]);
      setPosts([]);
      setEvents([]);
      setPendingRequests([]);
      router.push('/feed');
    } catch (err: any) {
      alert(err?.message ?? 'Failed to leave group');
    }
  };

  const handleMembersClick = () => {
    setShowMembersModal(true);
  };

  const handleInvite = async () => {
    try {
      let profileId = currentUserId;
      if (!profileId) {
        const me = await getProfile();
        profileId = (me as any)?.id ?? (me as any)?.user_id ?? (me as any)?.user?.id;
      }
      if (!profileId) {
        alert('Unable to get your profile. Please try logging in again.');
        return;
      }

      const userProfile = await getUserProfile(profileId);

      if (userProfile.privateProfile) {
        alert('Unable to load followers from a private profile.');
        setFollowers([]);
        setShowInviteModal(true);
        return;
      }

      const memberIds = new Set((members ?? []).map((m: any) => m.user_id));
      const followersList = Array.isArray(userProfile.followers) ? userProfile.followers as FollowerUser[] : [];

      const inviteCandidates = followersList.filter((f) => f.user_id && !memberIds.has(f.user_id));

      if (followersList.length === 0) {
        setFollowers([]);
        setInviteMessage('You don\'t have any followers yet. Only your followers can be invited to this group.');
        setShowInviteModal(true);
        return;
      }

      setFollowers(inviteCandidates);
      setInviteMessage(inviteCandidates.length === 0 ? 'All your followers are already members of this group.' : null);
      setShowInviteModal(true);
    } catch (err: any) {
      console.error('Failed to load followers:', err);
      alert(err?.message || 'Failed to load followers. Please try again.');
    }
  };

  const handleInviteUser = async (userId: string) => {
    if (!groupId) return;
    setInviteLoading(userId);
    try {
      await inviteToGroup(groupId, userId);
      // Optionally remove the invited user from the list or show success
      setFollowers((prev) => prev.filter((f) => f.user_id !== userId));
    } catch (err: any) {
      console.error('Failed to invite user:', err);
      alert(err?.message || 'Failed to invite user');
    } finally {
      setInviteLoading(null);
    }
  };

  const handleJoin = () => {
    if (!groupId || joinLoading || joinPending) return;
    setJoinLoading(true);
    requestToJoinGroup(groupId)
      .then(() => {
        setJoinPending(true);
      })
      .catch((err: any) => {
        const message = err?.message ?? '';
        if (message.toLowerCase().includes('already') && message.toLowerCase().includes('pending')) {
          setJoinPending(true);
          return;
        }
        alert(message || 'Failed to send join request');
      })
      .finally(() => setJoinLoading(false));
  };

  const handleRequestAction = async (requestId: string, action: 'approve' | 'deny') => {
    setRequestActionId(requestId);
    try {
      if (action === 'approve') {
        await approveGroupRequest(requestId);
        setGroup((prev: any) => {
          if (!prev) return prev;
          const nextCount = typeof prev.member_count === 'number' ? prev.member_count + 1 : prev.member_count;
          return { ...prev, member_count: nextCount };
        });
      } else {
        await denyGroupRequest(requestId);
      }
      setPendingRequests((prev) => prev.filter((r) => r.id !== requestId));
    } catch (err: any) {
      alert(err?.message ?? `Failed to ${action} request`);
    } finally {
      setRequestActionId(null);
    }
  };


  if (loading) {
    return (
      <main className="min-h-screen bg-parea-white px-16 py-12">
        <div className="max-w-7xl mx-auto">
          <p className="text-regular text-parea-black">Loading group...</p>
        </div>
      </main>
    );
  }

  if (error || !group) {
    return (
      <main className="min-h-screen bg-parea-white px-16 py-12">
        <div className="max-w-7xl mx-auto">
          <p className="text-regular text-parea-black">{error ?? 'Group not found'}</p>
        </div>
      </main>
    );
  }

  const groupWrapData = {
    name: group.title ?? '',
    description: group.description ?? undefined,
    admin: group.owner_nickname ?? '',
    createdDate: formatDate(group.created_at),
    membersCount: group.member_count ?? 0,
  };

  return (
    <main className="min-h-screen bg-parea-white px-16 py-12">
      <div className="max-w-7xl mx-auto space-y-8">
        <GroupWrap
          group={groupWrapData}
          isMember={isMember}
          isPending={joinPending}
          onJoin={handleJoin}
          onInvite={handleInvite}
          onLeaveGroup={handleLeaveGroup}
          onMembersClick={isMember ? handleMembersClick : undefined}
        />

        {!isMember && (
          <div className="rounded border border-parea-black bg-parea-white p-4">
            <p className="text-regular text-parea-black">
              {joinPending
                ? 'Join request sent. Wait for the group owner to approve it.'
                : joinLoading
                  ? 'Sending join request...'
                  : 'You are not a member yet. Send a join request to access posts and events.'}
            </p>
          </div>
        )}

        {isOwner && (
          <section className="rounded border border-parea-black bg-parea-white p-4">
            <h3 className="text-small font-medium uppercase text-parea-black">Pending Join Requests</h3>
            {pendingRequests.length === 0 ? (
              <p className="mt-3 text-regular text-parea-black">No pending requests.</p>
            ) : (
              <ul className="mt-3 flex flex-col gap-3">
                {pendingRequests.map((request) => (
                  <li key={request.id} className="flex flex-wrap items-center gap-2">
                    <span className="text-regular text-parea-black">
                      {request.user_nickname || request.first_name || request.user_id}
                    </span>
                    <button
                      type="button"
                      onClick={() => { void handleRequestAction(request.id, 'approve'); }}
                      disabled={requestActionId === request.id}
                      className="rounded border border-parea-black bg-parea-yellow px-3 py-1 text-small font-medium uppercase text-parea-black hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      Accept
                    </button>
                    <button
                      type="button"
                      onClick={() => { void handleRequestAction(request.id, 'deny'); }}
                      disabled={requestActionId === request.id}
                      className="rounded border border-parea-black bg-parea-white px-3 py-1 text-small font-medium uppercase text-parea-black hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                      Decline
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </section>
        )}

        <div>
          <div className="flex items-center justify-between gap-6">
            <Tabs
              tabs={['Posts', 'Events']}
              defaultTab="Posts"
              onTabChange={(tab) => setActiveTab(tab)}
            />
            <div className="flex items-center gap-2">
              {isMember && activeTab === 'Posts' && (
                <Button
                  variant="primary"
                  size="lg"
                  onClick={() => setCreatePostOpen(true)}
                >
                  Create Post
                </Button>
              )}
              {isMember && activeTab === 'Events' && (
                <Button
                  variant="primary"
                  size="lg"
                  onClick={() => setCreateEventOpen(true)}
                >
                  Create Event
                </Button>
              )}
              {isMember && (
                <Button
                  variant="secondary"
                  size="lg"
                  onClick={() => {
                    window.dispatchEvent(new CustomEvent('openGroupChat', {
                      detail: { groupId, groupTitle: group?.title ?? 'Group Chat' },
                    }));
                  }}
                >
                  Group Chat
                </Button>
              )}
            </div>
          </div>

          <div className="mt-6">
            {activeTab === 'Posts' && (
              <>
                {postsLoading && <p className="text-regular text-parea-black">Loading posts...</p>}
                {!postsLoading && posts.length === 0 && (
                  <p className="text-regular text-parea-black">No posts yet.</p>
                )}
                {!postsLoading && posts.length > 0 && (
                  <div className="flex flex-col items-start gap-0">
                    {posts.map((post, index) => {
                      const meta = postMetaById[post.id] || {};
                      const rawImagePath = meta.imageUrl || post.thumbnail_url || post.image_url;
                      const imageUrl = getPostImageUrl(rawImagePath);
                      const commentCount =
                        typeof post.comment_count === 'number'
                          ? post.comment_count
                          : typeof meta.commentCount === 'number'
                            ? meta.commentCount
                            : Array.isArray(post.comments)
                              ? post.comments.length
                              : 0;

                      const avatarSrc = meta.avatarUrl || '/user-avatar-default.png';
                      const avatarAlt = post.nickname ?? 'Author';
                      const userName = (post.nickname ?? 'User').toUpperCase();

                      return (
                        <Card
                          key={post.id}
                          imageType="post"
                          imageSrc={imageUrl}
                          avatarSrc={avatarSrc}
                          avatarAlt={avatarAlt}
                          userName={userName}
                          userDate={formatDate(post.created_at)}
                          title={post.title}
                          content={post.content}
                          href={`/post/${post.id}`}
                          imagePriority={index === 0}
                          commentCount={commentCount}
                        />
                      );
                    })}
                  </div>
                )}
              </>
            )}

            {activeTab === 'Events' && (
              <>
                {eventsLoading && <p className="text-regular text-parea-black">Loading events...</p>}
                {!eventsLoading && events.length === 0 && (
                  <p className="text-regular text-parea-black">No events yet.</p>
                )}
                {!eventsLoading && events.length > 0 && (
                  <div className="flex flex-col items-start gap-0">
                    {events.map((ev) => (
                      (() => {
                        const creatorID = ev.creator_id ?? ev.creatorId;
                        const creatorMember = Array.isArray(members)
                          ? members.find((member: any) => {
                              const memberID = member?.user_id ?? member?.id;
                              return memberID === creatorID;
                            })
                          : undefined;
                        const creatorFromMember =
                          creatorMember?.nickname ||
                          [creatorMember?.first_name, creatorMember?.last_name].filter(Boolean).join(' ').trim();
                        const creatorName =
                          ev.creator_nickname ||
                          ev.creator_name ||
                          creatorFromMember ||
                          creatorID ||
                          'User';
                        const creatorAvatarSrc =
                          (typeof creatorID === 'string' && eventCreatorAvatarById[creatorID]) ||
                          '/user-avatar-default.png';

                        return (
                          <Card
                            key={ev.id ?? ev.event_id}
                            imageType="post"
                            hideImage
                            hideReactions
                            avatarSrc={creatorAvatarSrc}
                            avatarAlt={creatorName}
                            userName={String(creatorName).toUpperCase()}
                            userDate={formatDate(ev.event_time ?? ev.created_at)}
                            title={ev.title}
                            content={ev.description}
                            href={`/event/${ev.id ?? ev.event_id}`}
                            imagePriority={false}
                          />
                        );
                      })()
                    ))}
                  </div>
                )}
              </>
            )}

          </div>
        </div>
      </div>

      {createPostOpen && (
        <CreatePostModal
          isOpen={createPostOpen}
          onClose={() => setCreatePostOpen(false)}
          groupId={groupId}
          onSuccess={() => {
            setCreatePostOpen(false);
            getGroupPosts(groupId).then((list) => setPosts(Array.isArray(list) ? list : []));
          }}
        />
      )}

      {createEventOpen && (
        <CreateEventModal
          isOpen={createEventOpen}
          onClose={() => setCreateEventOpen(false)}
          groupId={groupId}
          onSuccess={() => {
            setCreateEventOpen(false);
            getGroupEvents(groupId).then((list) => setEvents(Array.isArray(list) ? list : []));
          }}
        />
      )}

      <FollowersModal
        isOpen={showInviteModal}
        onClose={() => { setShowInviteModal(false); setInviteMessage(null); }}
        heading="Invite to group"
        users={followers}
        onInvite={followers.length > 0 ? handleInviteUser : undefined}
        isActionLoading={inviteLoading}
        emptyMessage={inviteMessage ?? undefined}
      />

      <FollowersModal
        isOpen={showMembersModal}
        onClose={() => setShowMembersModal(false)}
        heading="Members"
        users={membersModalUsers}
      />
    </main>
  );
}
