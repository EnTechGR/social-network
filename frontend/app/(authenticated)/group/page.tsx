'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import Button from '@/components/ui/Button';
import CreateGroupModal from '@/components/ui/CreateGroupModal';
import {
  getAllGroups,
  getMyGroups,
  getMyGroupInvites,
  requestToJoinGroup,
} from '@/lib/api';

function formatDate(iso: string): string {
  if (!iso) return '';
  try {
    return new Date(iso).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
  } catch {
    return iso;
  }
}

export default function GroupPage() {
  const router = useRouter();
  const [groups, setGroups] = useState<any[]>([]);
  const [myGroups, setMyGroups] = useState<any[]>([]);
  const [invites, setInvites] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [pendingRequestIds, setPendingRequestIds] = useState<Set<string>>(new Set());
  const [createGroupOpen, setCreateGroupOpen] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    Promise.all([getAllGroups(), getMyGroups(), getMyGroupInvites()])
      .then(([allGroups, memberGroups, pendingInvites]) => {
        if (cancelled) return;
        setGroups(Array.isArray(allGroups) ? allGroups : []);
        setMyGroups(Array.isArray(memberGroups) ? memberGroups : []);
        setInvites(Array.isArray(pendingInvites) ? pendingInvites : []);
      })
      .catch((err: any) => {
        if (cancelled) return;
        setError(err?.message ?? 'Failed to load groups');
        setGroups([]);
        setMyGroups([]);
        setInvites([]);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const myGroupIds = useMemo(() => new Set(myGroups.map((g) => g.id)), [myGroups]);
  const invitedGroupIds = useMemo(() => new Set(invites.map((i) => i.group_id)), [invites]);

  const handleJoinRequest = async (groupId: string) => {
    setActionLoading(groupId);
    try {
      await requestToJoinGroup(groupId);
      setPendingRequestIds((prev) => new Set(prev).add(groupId));
    } catch (err: any) {
      const message = err?.message ?? '';
      if (message.toLowerCase().includes('already') && message.toLowerCase().includes('pending')) {
        setPendingRequestIds((prev) => new Set(prev).add(groupId));
        return;
      }
      alert(message || 'Failed to send join request');
    } finally {
      setActionLoading(null);
    }
  };

  return (
    <main className="min-h-screen bg-parea-white px-16 py-12">
      <div className="mx-auto max-w-7xl space-y-6">
        <div className="flex items-center justify-between gap-4">
          <h2 className="text-h3 text-parea-black">Browse Groups</h2>
          <Button variant="primary" size="lg" onClick={() => setCreateGroupOpen(true)}>
            Create Group
          </Button>
        </div>

        {loading && <p className="text-regular text-parea-black">Loading groups...</p>}
        {!loading && error && <p className="text-regular text-parea-black">{error}</p>}
        {!loading && !error && groups.length === 0 && (
          <p className="text-regular text-parea-black">No groups available yet.</p>
        )}

        {!loading && !error && groups.length > 0 && (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            {groups.map((group) => {
              const isMember = myGroupIds.has(group.id);
              const hasInvite = invitedGroupIds.has(group.id);
              const hasPending = pendingRequestIds.has(group.id);
              const disableJoin = isMember || hasInvite || hasPending || actionLoading === group.id;

              return (
                <article key={group.id} className="rounded border border-parea-black bg-parea-white p-5 shadow-[4px_4px_0_0_var(--parea-black)]">
                  <h3 className="text-lg font-semibold text-parea-black">{group.title ?? 'Group'}</h3>
                  <p className="mt-2 line-clamp-3 text-regular text-parea-black">
                    {group.description || 'No description'}
                  </p>

                  <div className="mt-3 text-small text-parea-black/80">
                    <p>Owner: {group.owner_nickname || 'Unknown'}</p>
                    <p>Members: {group.member_count ?? 0}</p>
                    <p>Created: {formatDate(group.created_at)}</p>
                  </div>

                  <div className="mt-4 flex flex-wrap items-center gap-2">
                    <Link href={`/group/${group.id}`} className="rounded border border-parea-black bg-parea-yellow px-3 py-1 text-small font-medium uppercase text-parea-black">
                      Open
                    </Link>
                    {isMember && <span className="text-small text-parea-black">Member</span>}
                    {!isMember && hasInvite && <span className="text-small text-parea-black">Invitation pending</span>}
                    {!isMember && !hasInvite && (
                      <button
                        type="button"
                        onClick={() => { void handleJoinRequest(group.id); }}
                        disabled={disableJoin}
                        className="rounded border border-parea-black bg-parea-white px-3 py-1 text-small font-medium uppercase text-parea-black hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
                      >
                        {hasPending ? 'Request sent' : actionLoading === group.id ? 'Sending...' : 'Request to join'}
                      </button>
                    )}
                  </div>
                </article>
              );
            })}
          </div>
        )}
      </div>

      {createGroupOpen && (
        <CreateGroupModal
          isOpen={createGroupOpen}
          onClose={() => setCreateGroupOpen(false)}
          onSuccess={(groupId) => {
            setCreateGroupOpen(false);
            router.push(`/group/${groupId}`);
          }}
        />
      )}
    </main>
  );
}
