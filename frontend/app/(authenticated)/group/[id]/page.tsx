'use client';

import { useParams } from 'next/navigation';
import { useState, useEffect } from 'react';
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

export default function GroupDetailPage() {
  const params = useParams();
  const groupId = typeof params?.id === 'string' ? params.id : '';
  const [group, setGroup] = useState<any | null>(null);
  const [isMember, setIsMember] = useState(false);
  const [posts, setPosts] = useState<any[]>([]);
  const [events, setEvents] = useState<any[]>([]);
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

  useEffect(() => {
    if (!groupId) {
      setError('Invalid group');
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    setError(null);
    getGroupById(groupId)
      .then((g) => {
        if (cancelled) return;
        setGroup(g);
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
      .then(() => {
        if (!cancelled) setIsMember(true);
      })
      .catch(() => {
        if (!cancelled) setIsMember(false);
      });
    return () => { cancelled = true; };
  }, [groupId, group]);

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

  const handleLeaveGroup = () => {
    setIsMember(false);
    setPosts([]);
    setEvents([]);
  };

  const handleMembersClick = () => {
    // TODO: open members modal if needed
  };

  const handleInvite = async () => {
    try {
      const profile = await getProfile();
      if (!profile?.id) return;
      const userProfile = await getUserProfile(profile.id);
      if (userProfile.privateProfile) return;
      const followersList = Array.isArray(userProfile.followers) ? userProfile.followers as FollowerUser[] : [];
      setFollowers(followersList);
      setShowInviteModal(true);
    } catch (err) {
      console.error('Failed to load followers:', err);
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
    // TODO: request to join
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
          onJoin={handleJoin}
          onInvite={handleInvite}
          onLeaveGroup={handleLeaveGroup}
          onMembersClick={handleMembersClick}
        />

        <div>
          <Tabs
            tabs={['Posts', 'Events']}
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
                    {posts.map((post, index) => (
                      <Card
                        key={post.id}
                        imageType="post"
                        imageSrc={post.image_url || post.thumbnail_url}
                        avatarSrc="/user-avatar-default.png"
                        avatarAlt={post.nickname ?? 'Author'}
                        userName={(post.nickname ?? 'User').toUpperCase()}
                        userDate={formatDate(post.created_at)}
                        title={post.title}
                        content={post.content}
                        href={`/post/${post.id}`}
                        imagePriority={index === 0}
                      />
                    ))}
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
                        imagePriority={false}
                      />
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
        onClose={() => setShowInviteModal(false)}
        heading="Members"
        users={followers}
        onInvite={handleInviteUser}
        isActionLoading={inviteLoading}
      />
    </main>
  );
}
