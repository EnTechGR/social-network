'use client';

import { useParams } from 'next/navigation';
import { useState, useEffect, useRef } from 'react';
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
  requestToJoinGroup,
  getPendingGroupRequests,
  approveGroupRequest,
  denyGroupRequest,
  getGroupChatMessages,
  sendGroupChatMessage,
  getPostImageUrl,
  type GroupChatMessage,
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

function formatDateTime(iso: string): string {
  if (!iso) return '';
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

function buildWebSocketURL(): string {
  const customURL = process.env.NEXT_PUBLIC_WS_URL;
  if (customURL) {
    return customURL;
  }

  const apiURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  try {
    const url = new URL(apiURL);
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
    url.pathname = '/ws';
    url.search = '';
    return url.toString();
  } catch {
    return 'ws://localhost:8080/ws';
  }
}

function sortChatMessages(messages: GroupChatMessage[]): GroupChatMessage[] {
  return [...messages].sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
}

function mergeGroupChatMessages(
  existing: GroupChatMessage[],
  incoming: GroupChatMessage[],
): GroupChatMessage[] {
  const byID = new Map<string, GroupChatMessage>();
  [...existing, ...incoming].forEach((message) => {
    byID.set(message.message_id, message);
  });
  return sortChatMessages(Array.from(byID.values()));
}

export default function GroupDetailPage() {
  const params = useParams();
  const groupId = typeof params?.id === 'string' ? params.id : '';
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
  const [followers, setFollowers] = useState<FollowerUser[]>([]);
  const [inviteLoading, setInviteLoading] = useState<string | null>(null);
  const [joinLoading, setJoinLoading] = useState(false);
  const [joinPending, setJoinPending] = useState(false);
  const [requestActionId, setRequestActionId] = useState<string | null>(null);
  const [chatMessages, setChatMessages] = useState<GroupChatMessage[]>([]);
  const [chatLoading, setChatLoading] = useState(false);
  const [chatInput, setChatInput] = useState('');
  const [chatSending, setChatSending] = useState(false);
  const [chatError, setChatError] = useState<string | null>(null);
  const chatBoxRef = useRef<HTMLDivElement | null>(null);

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
        setCurrentUserId(profile?.id ?? null);
        setIsOwner(Boolean(g?.owner_id && profile?.id && g.owner_id === profile.id));
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
    if (!groupId || !isMember || activeTab !== 'Chat') return;
    let cancelled = false;
    setChatLoading(true);
    setChatError(null);
    getGroupChatMessages(groupId, 200, 0)
      .then((messages) => {
        if (cancelled) return;
        setChatMessages(sortChatMessages(Array.isArray(messages) ? messages : []));
      })
      .catch((err: any) => {
        if (cancelled) return;
        setChatMessages([]);
        setChatError(err?.message ?? 'Failed to load group chat');
      })
      .finally(() => {
        if (!cancelled) setChatLoading(false);
      });
    return () => { cancelled = true; };
  }, [groupId, isMember, activeTab]);

  useEffect(() => {
    if (!groupId || !isMember) return;

    const wsURL = buildWebSocketURL();
    let isUnmounted = false;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
    let socket: WebSocket | null = null;

    const connect = () => {
      if (isUnmounted) {
        return;
      }

      try {
        socket = new WebSocket(wsURL);

        const processPayload = (raw: string) => {
          const payloads = raw
            .split('\n')
            .map((line) => line.trim())
            .filter((line) => line.length > 0);

          payloads.forEach((payload) => {
            try {
              const parsed = JSON.parse(payload) as { type?: string; data?: unknown };
              if (parsed.type !== 'group_chat' || !parsed.data || typeof parsed.data !== 'object') {
                return;
              }

              const message = parsed.data as Partial<GroupChatMessage>;
              if (
                typeof message.message_id !== 'string' ||
                typeof message.group_id !== 'string' ||
                typeof message.sender_id !== 'string' ||
                typeof message.sender_name !== 'string' ||
                typeof message.content !== 'string' ||
                typeof message.created_at !== 'string'
              ) {
                return;
              }

              if (message.group_id !== groupId) {
                return;
              }

              setChatMessages((prev) => mergeGroupChatMessages(prev, [message as GroupChatMessage]));
            } catch (error) {
              console.error('Failed to parse group chat websocket payload:', error);
            }
          });
        };

        socket.onmessage = (event: MessageEvent) => {
          if (typeof event.data === 'string') {
            processPayload(event.data);
            return;
          }

          if (event.data instanceof Blob) {
            event.data
              .text()
              .then(processPayload)
              .catch((error) => {
                console.error('Failed to decode group chat websocket blob:', error);
              });
          }
        };

        socket.onerror = () => {
          socket?.close();
        };

        socket.onclose = () => {
          if (isUnmounted) {
            return;
          }
          reconnectTimer = setTimeout(connect, 3000);
        };
      } catch (error) {
        console.error('Failed to open group chat websocket:', error);
        reconnectTimer = setTimeout(connect, 3000);
      }
    };

    connect();

    return () => {
      isUnmounted = true;
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
      }
      if (socket) {
        socket.onclose = null;
        socket.close();
      }
    };
  }, [groupId, isMember]);

  useEffect(() => {
    if (activeTab !== 'Chat') return;
    if (!chatBoxRef.current) return;
    chatBoxRef.current.scrollTop = chatBoxRef.current.scrollHeight;
  }, [chatMessages, activeTab]);

  const handleLeaveGroup = () => {
    alert('Leaving groups is not available in this build yet.');
  };

  const handleMembersClick = () => {
    // TODO: open members modal if needed
  };

  const handleInvite = async () => {
    try {
      const profileId = currentUserId ?? (await getProfile())?.id;
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
      
      if (followersList.length === 0) {
        alert('You don\'t have any followers yet. Only your followers can be invited to this group.');
        return;
      }
      
      const inviteCandidates = followersList.filter((f) => f.user_id && !memberIds.has(f.user_id));
      
      if (inviteCandidates.length === 0) {
        alert('All of your followers are already members of this group.');
        return;
      }
      
      setFollowers(inviteCandidates);
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

  const handleSendGroupChatMessage = async () => {
    if (!groupId || !isMember || chatSending) return;
    const content = chatInput.trim();
    if (!content) return;

    setChatSending(true);
    setChatError(null);
    try {
      const message = await sendGroupChatMessage(groupId, content);
      setChatMessages((prev) => mergeGroupChatMessages(prev, [message]));
      setChatInput('');
    } catch (err: any) {
      setChatError(err?.message ?? 'Failed to send message');
    } finally {
      setChatSending(false);
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
          onMembersClick={handleMembersClick}
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
          <Tabs
            tabs={['Posts', 'Events', 'Chat']}
            defaultTab="Posts"
            onTabChange={(tab) => setActiveTab(tab)}
          />

          <div className="mt-6">
            {activeTab === 'Posts' && (
              <>
                {isMember && (
                  <div className="mb-4">
                    <Button
                      variant="primary"
                      size="lg"
                      onClick={() => setCreatePostOpen(true)}
                    >
                      Create Post
                    </Button>
                  </div>
                )}
                {postsLoading && <p className="text-regular text-parea-black">Loading posts...</p>}
                {!postsLoading && posts.length === 0 && (
                  <p className="text-regular text-parea-black">No posts yet.</p>
                )}
                {!postsLoading && posts.length > 0 && (
                  <div className="flex flex-col items-start gap-0">
                    {posts.map((post, index) => {
                      const rawImagePath = post.thumbnail_url || post.image_url;
                      const imageUrl = getPostImageUrl(rawImagePath);

                      return (
                        <Card
                          key={post.id}
                          imageType="post"
                          imageSrc={imageUrl}
                          avatarSrc="/user-avatar-default.png"
                          avatarAlt={post.nickname ?? 'Author'}
                          userName={(post.nickname ?? 'User').toUpperCase()}
                          userDate={formatDate(post.created_at)}
                          title={post.title}
                          content={post.content}
                          href={`/post/${post.id}`}
                          imagePriority={index === 0}
                        />
                      );
                    })}
                  </div>
                )}
              </>
            )}

            {activeTab === 'Events' && (
              <>
                {isMember && (
                  <div className="mb-4">
                    <Button
                      variant="primary"
                      size="lg"
                      onClick={() => setCreateEventOpen(true)}
                    >
                      Create Event
                    </Button>
                  </div>
                )}
                {eventsLoading && <p className="text-regular text-parea-black">Loading events...</p>}
                {!eventsLoading && events.length === 0 && (
                  <p className="text-regular text-parea-black">No events yet.</p>
                )}
                {!eventsLoading && events.length > 0 && (
                  <div className="flex flex-col items-start gap-0">
                    {events.map((ev) => (
                      <Card
                        key={ev.id ?? ev.event_id}
                        imageType="event"
                        avatarSrc="/user-avatar-default.png"
                        avatarAlt={ev.creator_nickname ?? 'Creator'}
                        userName={(ev.creator_nickname ?? 'User').toUpperCase()}
                        userDate={formatDate(ev.event_time ?? ev.created_at)}
                        title={ev.title}
                        content={ev.description}
                        href={`/event/${ev.id ?? ev.event_id}`}
                        imagePriority={false}
                      />
                    ))}
                  </div>
                )}
              </>
            )}

            {activeTab === 'Chat' && (
              <>
                {!isMember ? (
                  <p className="text-regular text-parea-black">
                    Join the group to access the group chat room.
                  </p>
                ) : (
                  <div className="flex flex-col rounded border border-parea-black bg-parea-white">
                    <div ref={chatBoxRef} className="h-[420px] overflow-y-auto border-b border-parea-black p-4">
                      {chatLoading ? (
                        <p className="text-regular text-parea-black">Loading chat...</p>
                      ) : chatMessages.length === 0 ? (
                        <p className="text-regular text-parea-black">No messages yet.</p>
                      ) : (
                        <div className="flex flex-col gap-2">
                          {chatMessages.map((message) => {
                            const isMine = currentUserId === message.sender_id;
                            return (
                              <div
                                key={message.message_id}
                                className={`max-w-[85%] rounded border border-parea-black px-3 py-2 ${
                                  isMine ? 'ml-auto bg-parea-yellow/40' : 'mr-auto bg-parea-white'
                                }`}
                              >
                                {!isMine && (
                                  <p className="text-[11px] uppercase text-parea-black/70">
                                    {message.sender_name}
                                  </p>
                                )}
                                <p className="whitespace-pre-wrap break-words text-regular text-parea-black">
                                  {message.content}
                                </p>
                                <p className="mt-1 text-[10px] text-parea-black/60">
                                  {formatDateTime(message.created_at)}
                                </p>
                              </div>
                            );
                          })}
                        </div>
                      )}
                    </div>

                    <div className="p-4">
                      {chatError && (
                        <p className="mb-2 text-small text-parea-black">{chatError}</p>
                      )}
                      <div className="flex gap-2">
                        <input
                          type="text"
                          value={chatInput}
                          onChange={(event) => setChatInput(event.target.value)}
                          onKeyDown={(event) => {
                            if (event.key === 'Enter' && !event.shiftKey) {
                              event.preventDefault();
                              void handleSendGroupChatMessage();
                            }
                          }}
                          placeholder="Type a message..."
                          className="flex-1 rounded border border-parea-black bg-parea-white px-3 py-2 text-regular text-parea-black focus:outline-none"
                        />
                        <button
                          type="button"
                          onClick={() => {
                            void handleSendGroupChatMessage();
                          }}
                          disabled={chatSending || chatInput.trim().length === 0}
                          className="rounded border border-parea-black bg-parea-yellow px-4 py-2 text-small font-medium uppercase text-parea-black hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
                        >
                          Send
                        </button>
                      </div>
                    </div>
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
        onClose={() => setShowInviteModal(false)}
        heading="Members"
        users={followers}
        onInvite={handleInviteUser}
        isActionLoading={inviteLoading}
      />
    </main>
  );
}
