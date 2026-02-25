'use client';

import { useState } from 'react';
import Button from './Button';
import IconButton from './IconButtons';
import Image from 'next/image';
import { createGroupEvent } from '@/lib/api';
import { useRouter } from 'next/navigation';

interface CreateEventModalProps {
  isOpen: boolean;
  onClose: () => void;
  groupId: string;
  onSuccess?: () => void;
}

export default function CreateEventModal({ isOpen, onClose, groupId, onSuccess }: CreateEventModalProps) {
  const router = useRouter();
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [eventTime, setEventTime] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleClose = () => {
    setTitle('');
    setDescription('');
    setEventTime('');
    setError(null);
    onClose();
  };

  const handleSubmit = async () => {
    if (!title.trim()) return;
    if (!eventTime) {
      setError('Event day/time is required.');
      return;
    }
    setError(null);
    setIsSubmitting(true);
    try {
      const result = await createGroupEvent(groupId, {
        title: title.trim(),
        description: description.trim(),
        event_time: new Date(eventTime).toISOString(),
      });

      const newEventId =
        result?.id ??
        result?.event_id ??
        result?.event?.id ??
        result?.event?.event_id;

      handleClose();

      if (newEventId) {
        router.push(`/event/${newEventId}`);
      } else {
        onSuccess?.();
      }
    } catch (err: any) {
      setError(err?.message ?? 'Failed to create event.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      onClick={handleClose}
    >
      <div className="absolute inset-0 bg-black/50" />
      <div className="relative w-full max-w-145 bg-white border border-parea-black shadow-[8px_8px_0_0_#000] mx-4" onClick={(e) => e.stopPropagation()}>
        <div className="relative h-17 border-b overflow-hidden bg-parea-white">
          <Image
            src="/modal-header-pattern.svg"
            alt=""
            fill
            className="object-cover"
          />
          <IconButton
            variant="close"
            onClick={handleClose}
            aria-label="Close modal"
            className="absolute top-4 right-8 z-10"
          />
        </div>
        <div className="p-8 min-h-129 flex flex-col">
          <h4 className="mb-4">Create event</h4>
          <div className="mb-6">
            <label className="label block mb-2">TITLE</label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full p-3 bg-parea-white border border-parea-black focus:outline-none"
            />
          </div>
          <div className="mb-6">
            <label className="label block mb-2">DESCRIPTION</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Event details"
              className="w-full min-h-24 p-4 bg-parea-white border border-parea-black focus:outline-none resize-none"
              style={{ fontFamily: 'var(--font-body), system-ui, sans-serif' }}
            />
          </div>
          <div className="mb-6">
            <label className="label block mb-2">DATE & TIME</label>
            <input
              type="datetime-local"
              value={eventTime}
              onChange={(e) => setEventTime(e.target.value)}
              className="w-full p-3 bg-parea-white border border-parea-black focus:outline-none"
            />
          </div>
          {error && <p className="text-sm text-red-600 mb-4">{error}</p>}
          <div className="flex justify-end">
            <Button
              variant="primary"
              size="lg"
              onClick={handleSubmit}
              disabled={isSubmitting || !title.trim()}
            >
              {isSubmitting ? 'SUBMITTING...' : 'SUBMIT'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
