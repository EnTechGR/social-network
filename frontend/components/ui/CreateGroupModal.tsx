'use client';

import { useState } from 'react';
import Button from './Button';
import IconButton from './IconButtons';
import Image from 'next/image';
import { createGroup } from '@/lib/api';

interface CreateGroupModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: (groupId: string) => void;
}

export default function CreateGroupModal({ isOpen, onClose, onSuccess }: CreateGroupModalProps) {
  const [name, setName] = useState('');
  const [bio, setBio] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const handleClose = () => {
    setName('');
    setBio('');
    setError(null);
    onClose();
  };

  const handleSubmit = async () => {
    if (!name.trim()) return;
    setError(null);
    setIsSubmitting(true);
    try {
      const res = await createGroup({ title: name.trim(), description: bio.trim() });
      const groupId = res?.group?.id;
      handleClose();
      if (groupId) onSuccess?.(groupId);
    } catch (err: any) {
      setError(err?.message ?? 'Failed to create group.');
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
          <h4 className="mb-4">Create group</h4>
          <div className="mb-6">
            <label className="label block mb-2">GROUP NAME</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full p-3 bg-parea-white border border-parea-black focus:outline-none"
            />
          </div>
          <div className="mb-6">
            <label className="label block mb-2">GROUP BIO</label>
            <textarea
              value={bio}
              onChange={(e) => setBio(e.target.value)}
              placeholder="What is this group about?"
              className="w-full min-h-24 p-4 bg-parea-white border border-parea-black focus:outline-none resize-none"
              style={{ fontFamily: 'var(--font-body), system-ui, sans-serif' }}
            />
          </div>
          {error && <p className="text-sm text-red-600 mb-4">{error}</p>}
          <div className="flex justify-end">
            <Button
              variant="primary"
              size="lg"
              onClick={handleSubmit}
              disabled={isSubmitting || !name.trim()}
            >
              {isSubmitting ? 'SUBMITTING...' : 'SUBMIT'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
