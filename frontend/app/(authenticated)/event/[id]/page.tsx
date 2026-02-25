'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import EventDetailFrame from '@/components/ui/EventDetailFrame';
import type { RsvpOption } from '@/components/ui/EventInfoBar';
import FollowersModal, { type FollowerUser } from '@/components/ui/FollowersModal';
import {
  getEventById,
  voteOnEvent,
  getProfile,
  getUserProfile,
  getGroupMembers,
  inviteToGroup,
} from '@/lib/api';

function formatDateParts(iso: string): { date: string; time: string } {
  if (!iso) return { date: '', time: '' };
  try {
    const date = new Date(iso);
    return {
      date: date.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' }).toUpperCase(),
      time: date.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' }),
    };
  } catch {
    return { date: iso, time: '' };
  }
}

function choiceToUi(choice?: string | null): RsvpOption {
  switch ((choice || '').toLowerCase()) {
    case 'going':
      return 'GOING';
    case 'not going':
      return 'NOT GOING';
    case 'maybe':
      return 'MAYBE';
    default:
      return 'RSVP';
  }
}

function uiToChoice(choice: RsvpOption): 'going' | 'not going' | 'maybe' | null {
  switch (choice) {
    case 'GOING':
      return 'going';
    case 'NOT GOING':
      return 'not going';
    case 'MAYBE':
      return 'maybe';
    default:
      return null;
  }
}

export default function EventPage() {
  const params = useParams();
  const eventId = typeof params?.id === 'string' ? params.id : '';
  const [eventData, setEventData] = useState<any | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [rsvpValue, setRsvpValue] = useState<RsvpOption>('RSVP');
  const [isVoting, setIsVoting] = useState(false);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [inviteCandidates, setInviteCandidates] = useState<FollowerUser[]>([]);
  const [inviteLoading, setInviteLoading] = useState<string | null>(null);

  const loadEvent = async () => {
    if (!eventId) return;
    const data = await getEventById(eventId);
    setEventData(data);
    setRsvpValue(choiceToUi(data?.user_vote));
  };

  useEffect(() => {
    if (!eventId) {
      setError('Invalid event');
      setLoading(false);
      return;
    }
    let cancelled = false;
    setLoading(true);
    setError(null);
    loadEvent()
      .catch((err: any) => {
        if (!cancelled) setError(err?.message ?? 'Failed to load event');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [eventId]);

  const handleRsvpSelect = async (option: RsvpOption) => {
    const choice = uiToChoice(option);
    if (!choice || isVoting || !eventId) return;
    setIsVoting(true);
    try {
      await voteOnEvent(eventId, choice);
      await loadEvent();
    } catch (err: any) {
      alert(err?.message ?? 'Failed to submit vote');
    } finally {
      setIsVoting(false);
    }
  };

  const handleInviteClick = async () => {
    if (!eventData?.group_id) {
      alert('This event is not linked to a group.');
      return;
    }

    const groupId = eventData.group_id as string;

    try {
      const me = await getProfile();
      const profileId = me?.id;

      if (!profileId) {
        alert('Unable to get your profile. Please try logging in again.');
        return;
      }

      const userProfile = await getUserProfile(profileId);

      if (userProfile.privateProfile) {
        alert('Unable to load followers from a private profile.');
        setInviteCandidates([]);
        setShowInviteModal(true);
        return;
      }

      const members = await getGroupMembers(groupId);
      const memberIds = new Set(
        (Array.isArray(members) ? members : []).map((m: any) => m.user_id),
      );

      const followersList = Array.isArray(userProfile.followers)
        ? (userProfile.followers as FollowerUser[])
        : [];

      if (followersList.length === 0) {
        alert("You don't have any followers yet. Only your followers can be invited to this group.");
        return;
      }

      const candidates = followersList.filter(
        (f) => f.user_id && !memberIds.has(f.user_id),
      );

      if (candidates.length === 0) {
        alert('All of your followers are already members of this group.');
        return;
      }

      setInviteCandidates(candidates);
      setShowInviteModal(true);
    } catch (err: any) {
      console.error('Failed to load followers for invite:', err);
      alert(err?.message || 'Failed to load followers. Please try again.');
    }
  };

  const handleInviteUser = async (userId: string) => {
    if (!eventData?.group_id) return;
    const groupId = eventData.group_id as string;

    setInviteLoading(userId);
    try {
      await inviteToGroup(groupId, userId);
      setInviteCandidates((prev) => prev.filter((u) => u.user_id !== userId));
    } catch (err: any) {
      console.error('Failed to invite user:', err);
      alert(err?.message || 'Failed to invite user');
    } finally {
      setInviteLoading(null);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-parea-white px-16 py-12">
        <p className="text-regular text-parea-black">Loading event...</p>
      </div>
    );
  }

  if (error || !eventData) {
    return (
      <div className="min-h-screen bg-parea-white px-16 py-12">
        <p className="text-regular text-parea-black">{error ?? 'Event not found'}</p>
      </div>
    );
  }

  const { date, time } = formatDateParts(eventData.event_time);
  const goingOption = Array.isArray(eventData.options)
    ? eventData.options.find((o: any) => (o?.label || '').toLowerCase() === 'going')
    : null;
  const goingCount = goingOption?.vote_count ?? 0;

  return (
    <div className="min-h-screen bg-parea-white px-16 py-12">
      <div className="w-full">
        <EventDetailFrame
          title={eventData.title ?? 'Event'}
          imageSrc="/eventDefaultImage.png"
          location={eventData.group_title || 'GROUP EVENT'}
          goingCount={goingCount}
          date={date}
          time={time}
          eventText={eventData.description || ''}
          rsvpValue={rsvpValue}
          onRsvpSelect={handleRsvpSelect}
          onInviteClick={handleInviteClick}
        />
        {isVoting && (
          <p className="mt-4 text-small text-parea-black">Updating your RSVP...</p>
        )}
        <FollowersModal
          isOpen={showInviteModal}
          onClose={() => setShowInviteModal(false)}
          heading="Members"
          users={inviteCandidates}
          onInvite={handleInviteUser}
          isActionLoading={inviteLoading}
        />
      </div>
    </div>
  );
}
