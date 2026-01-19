'use client';

import { useState, useRef, useEffect } from 'react';
import Button from './Button';
import IconButton from './IconButtons';
import { ImageUp } from 'lucide-react';

interface CreatePostModalProps {
  isOpen: boolean;
  onClose: () => void;
  preview?: boolean;
}

export default function CreatePostModal({ isOpen, onClose, preview = false }: CreatePostModalProps) {
  const [title, setTitle] = useState('');
  const [details, setDetails] = useState('');
  const [visibility, setVisibility] = useState<'PUBLIC' | 'FOLLOWERS' | 'PRIVATE'>('PUBLIC');
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsDropdownOpen(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  if (!isOpen && !preview) return null;

  const handleSubmit = () => {
    console.log({ title, details, visibility });
    onClose();
  };

  const modalContent = (
    <div className="relative w-full max-w-145 bg-parea-white border border-parea-black">
      {/* Header */}
      <div className="relative h-18 border-b">
        {/* Close Button */}
        <IconButton
          variant="close"
          onClick={onClose}
          aria-label="Close modal"
          className="absolute top-4 right-4"
        />
      </div>

      {/* Content */}
      <div className="p-8">
        {/* Header Row */}
        <div className="flex items-center justify-between mb-4">
          <h4>Create post</h4>

          {/* Visibility Dropdown */}
          <div ref={dropdownRef} className="relative">
            <button
              type="button"
              onClick={() => setIsDropdownOpen(!isDropdownOpen)}
              className="flex items-center justify-between w-34 px-5 py-2 bg-white border border-[#222] cursor-pointer shadow-[0.25rem_0.25rem_0_0_#000]"
              style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace' }}
            >
              <span className="text-base font-medium uppercase tracking-[-0.01em] leading-relaxed">
                {visibility}
              </span>
              <svg
                width="12"
                height="7"
                viewBox="0 0 12 7"
                fill="none"
                className={`transition-transform duration-200 ${isDropdownOpen ? 'rotate-180' : ''}`}
              >
                <path
                  d="M1 1L6 6L11 1"
                  stroke="#000000"
                  strokeWidth="1.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </button>

            {isDropdownOpen && (
              <div className="absolute top-full left-0 w-34 bg-white border border-[#222] border-t-0 z-10 shadow-[0.25rem_0.25rem_0_0_#000]">
                <button
                  type="button"
                  onClick={() => {
                    setVisibility('PUBLIC');
                    setIsDropdownOpen(false);
                  }}
                  className={`block w-full px-5 py-2 text-left text-base font-medium uppercase tracking-[-0.01em] leading-relaxed hover:bg-parea-grey transition-colors cursor-pointer ${visibility === 'PUBLIC' ? 'bg-parea-grey' : ''
                    }`}
                  style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace' }}
                >
                  PUBLIC
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setVisibility('FOLLOWERS');
                    setIsDropdownOpen(false);
                  }}
                  className={`block w-full px-5 py-2 text-left text-base font-medium uppercase tracking-[-0.01em] leading-relaxed hover:bg-parea-grey transition-colors cursor-pointer ${visibility === 'FOLLOWERS' ? 'bg-parea-grey' : ''
                    }`}
                  style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace' }}
                >
                  FOLLOWERS
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setVisibility('PRIVATE');
                    setIsDropdownOpen(false);
                  }}
                  className={`block w-full px-5 py-2 text-left text-base font-medium uppercase tracking-[-0.01em] leading-relaxed hover:bg-parea-grey transition-colors cursor-pointer ${visibility === 'PRIVATE' ? 'bg-parea-grey' : ''
                    }`}
                  style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace' }}
                >
                  PRIVATE
                </button>
              </div>
            )}
          </div>
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
            className="w-full p-3 bg-parea-grey border border-parea-black focus:outline-none"
          />
        </div>

        {/* Details Field */}
        <div className="mb-6">
          <label className="label block mb-2">
            DETAILS
          </label>
          <div className="relative">
            <textarea
              value={details}
              onChange={(e) => setDetails(e.target.value)}
              placeholder="What are you thinking?"
              className="w-full h-46 p-4 bg-parea-grey border border-parea-black focus:outline-none resize-none"
              style={{ fontFamily: 'var(--font-body), system-ui, sans-serif' }}
            />
            {/* Image Upload Button */}
            <button
              type="button"
              className="absolute bottom-4 right-4 hover:bg-parea-black/10 rounded transition-colors"
            >
              <ImageUp size={32} stroke="#000" strokeWidth={1.5} />
            </button>
          </div>
        </div>

        {/* Submit Button */}
        <div className="flex justify-end">
          <Button variant="primary" size="lg" onClick={handleSubmit}>
            SUBMIT
          </Button>
        </div>
      </div>
    </div>
  );

  if (preview) {
    return modalContent;
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      onClick={onClose}
    >
      <div className="absolute inset-0 bg-black/50" />
      <div onClick={(e) => e.stopPropagation()}>
        {modalContent}
      </div>
    </div>
  );
}
