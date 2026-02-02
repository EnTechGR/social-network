'use client';

import { useState, useRef } from 'react';
import Button from './Button';
import IconButton from './IconButtons';
import Avatar from './Avatar';
import DropdownButton from './DropdownButton';
import { ImagePlus } from 'lucide-react';
import Image from 'next/image';

const VISIBILITY_OPTIONS = ['PUBLIC', 'FOLLOWERS', 'PRIVATE'] as const;

interface Follower {
  id: string;
  name: string;
  avatarUrl?: string;
}

interface CreatePostModalProps {
  isOpen: boolean;
  onClose: () => void;
  preview?: boolean;
  /** List of followers to select from when visibility is FOLLOWERS */
  followers?: Follower[];
}

// Mock followers for preview/demo
const mockFollowers: Follower[] = [
  { id: '1', name: 'Alex Donham', avatarUrl: '/test-avatar.png' },
  { id: '2', name: 'Anette Black', avatarUrl: '/test-avatar.png' },
  { id: '3', name: 'Mario Salvante', avatarUrl: '/test-avatar.png' },
  { id: '4', name: 'Karen Hills', avatarUrl: '/test-avatar.png' },
  { id: '5', name: 'Jacob Jones', avatarUrl: '/test-avatar.png' },
  { id: '6', name: 'Ammy Stones', avatarUrl: '/test-avatar.png' },
];

export default function CreatePostModal({ isOpen, onClose, preview = false, followers = mockFollowers }: CreatePostModalProps) {
  const [title, setTitle] = useState('');
  const [details, setDetails] = useState('');
  const [visibility, setVisibility] = useState<'PUBLIC' | 'FOLLOWERS' | 'PRIVATE'>('PUBLIC');
  const [step, setStep] = useState(1);
  const [selectedFollowers, setSelectedFollowers] = useState<Set<string>>(new Set());
  const [isAtBottom, setIsAtBottom] = useState(false);
  const [uploadedImage, setUploadedImage] = useState<{ file: File; preview: string } | null>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleImageUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      const preview = URL.createObjectURL(file);
      setUploadedImage({ file, preview });
    }
  };

  const handleRemoveImage = () => {
    if (uploadedImage) {
      URL.revokeObjectURL(uploadedImage.preview);
      setUploadedImage(null);
    }
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  // Check if scrolled to bottom of followers list
  const handleScroll = () => {
    if (listRef.current) {
      const { scrollTop, scrollHeight, clientHeight } = listRef.current;
      const atBottom = scrollTop + clientHeight >= scrollHeight - 10; // 10px threshold
      setIsAtBottom(atBottom);
    }
  };

  if (!isOpen && !preview) return null;

  const toggleFollowerSelection = (followerId: string) => {
    setSelectedFollowers(prev => {
      const newSet = new Set(prev);
      if (newSet.has(followerId)) {
        newSet.delete(followerId);
      } else {
        newSet.add(followerId);
      }
      return newSet;
    });
  };

  const handleNext = () => {
    if (visibility === 'FOLLOWERS') {
      setStep(2);
    } else {
      handleSubmit();
    }
  };

  const handleBack = () => {
    setStep(1);
  };

  const resetModal = () => {
    setStep(1);
    setSelectedFollowers(new Set());
    handleRemoveImage();
  };

  const handleClose = () => {
    resetModal();
    onClose();
  };

  const handleSubmit = () => {
    console.log({ title, details, visibility, selectedFollowers: Array.from(selectedFollowers) });
    resetModal();
    onClose();
  };

  // Step 1: Create Post Form
  const step1Content = (
    <div className="relative w-full max-w-145 bg-white border border-parea-black shadow-[8px_8px_0_0_#000]">
      {/* Header */}
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

      {/* Content */}
      <div className="p-8 min-h-129 flex flex-col">
        {/* Header Row */}
        <div className="flex items-center justify-between mb-4">
          <h4>Create post</h4>

          {/* Visibility Dropdown */}
          <DropdownButton
            options={[...VISIBILITY_OPTIONS]}
            value={visibility}
            onValueChange={(v) => setVisibility(v as 'PUBLIC' | 'FOLLOWERS' | 'PRIVATE')}
            aria-label="Post visibility"
          />
        </div>

        {/* Title Field */}
        <div className="mb-6">
          <label className="label block mb-2">
            TITLE
          </label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="w-full p-3 bg-parea-white border border-parea-black focus:outline-none"
          />
        </div>

        {/* Details Field */}
        <div className="mb-6 flex-1 flex flex-col">
          <label className="label block mb-2">
            DETAILS
          </label>
          <div className="relative">
            <textarea
              value={details}
              onChange={(e) => setDetails(e.target.value)}
              placeholder="What are you thinking?"
              className="w-full h-full min-h-35 p-4 bg-parea-white border border-parea-black focus:outline-none resize-none"
              style={{ fontFamily: 'var(--font-body), system-ui, sans-serif' }}
            />
            {/* Hidden file input */}
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              onChange={handleImageUpload}
              className="hidden"
            />
            {/* Image Upload Button */}
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              className="absolute bottom-4 right-3 hover:bg-parea-black/10 rounded transition-colors"
            >
              <ImagePlus size={24} stroke="#000" strokeWidth={1.5} />
            </button>
          </div>

          {/* Uploaded Image Preview */}
          {uploadedImage && (
            <div className="flex items-center gap-3 mt-3 p-3 border border-parea-border bg-parea-white">
              <Image
                src={uploadedImage.preview}
                alt="Uploaded preview"
                width={40}
                height={40}
                className="object-cover"
              />
              <span className="flex-1 text-sm text-parea-black/60 truncate">
                {uploadedImage.file.name}
              </span>
              <IconButton
                variant="close"
                size="sm"
                transparent
                onClick={handleRemoveImage}
                aria-label="Remove image"
              />
            </div>
          )}
        </div>

        {/* Submit/Next Button - fixed height container to prevent layout shift */}
        <div className="flex justify-end h-12 items-center">
          {visibility === 'FOLLOWERS' ? (
            <IconButton
              variant="arrow-right"
              text="NEXT"
              onClick={handleNext}
              aria-label="Next step"
            />
          ) : (
            <Button variant="primary" size="lg" onClick={handleSubmit}>
              SUBMIT
            </Button>
          )}
        </div>
      </div>
    </div>
  );

  // Step 2: Follower Selection
  const step2Content = (
    <div className="relative w-full max-w-145 bg-white border border-parea-black shadow-[8px_8px_0_0_#000]">
      {/* Header */}
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

      {/* Content */}
      <div className="p-8 flex flex-col h-129">
        {/* Heading */}
        <div className="mb-4">
          <h4 className="text-2xl font-bold leading-normal">
            Choose who can see your post.
          </h4>
        </div>

        {/* Followers List with fade gradient */}
        <div className="relative flex-1 min-h-0">
          <div
            ref={listRef}
            onScroll={handleScroll}
            className="h-full overflow-y-auto pb-8"
          >
            {followers.map((follower) => (
              <div
                key={follower.id}
                className="flex items-center justify-between h-14 px-2"
              >
                <div className="flex items-center gap-3">
                  <Avatar
                    size="sm"
                    type={follower.avatarUrl ? 'image' : 'user'}
                    src={follower.avatarUrl}
                    alt={follower.name}
                  />
                  <span
                    className="text-base font-medium uppercase tracking-[-0.01em] leading-relaxed"
                    style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace' }}
                  >
                    {follower.name}
                  </span>
                </div>
                <button
                  type="button"
                  onClick={() => toggleFollowerSelection(follower.id)}
                  className={`text-base font-medium uppercase tracking-[-0.01em] leading-relaxed cursor-pointer border-b-2 border-parea-black text-parea-black
                    }`}
                  style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace' }}
                >
                  {selectedFollowers.has(follower.id) ? 'SELECTED' : 'SELECT'}
                </button>
              </div>
            ))}
          </div>
          {/* Fade gradient at bottom to indicate scrollability - fades out when at bottom */}
          <div
            className={`absolute bottom-0 left-0 right-0 h-20 pointer-events-none transition-opacity duration-100 ${isAtBottom ? 'opacity-0' : 'opacity-100'
              }`}
            style={{
              background: 'linear-gradient(to bottom, transparent 0%, #fff 100%)'
            }}
          />
        </div>

        {/* Navigation */}
        <div className="flex items-center justify-between pt-4">
          <IconButton
            variant="arrow-left"
            text="BACK"
            onClick={handleBack}
            aria-label="Go back"
          />
          <IconButton
            variant="arrow-right"
            text="NEXT"
            onClick={handleSubmit}
            aria-label="Submit post"
          />
        </div>
      </div>
    </div>
  );

  const modalContent = step === 1 ? step1Content : step2Content;

  if (preview) {
    return modalContent;
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      onClick={handleClose}
    >
      <div className="absolute inset-0 bg-black/50" />
      <div onClick={(e) => e.stopPropagation()}>
        {modalContent}
      </div>
    </div>
  );
}
