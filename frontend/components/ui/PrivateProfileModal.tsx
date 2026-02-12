'use client';

import { useState, useEffect } from 'react';
import Button from './Button';
import IconButton from './IconButtons';

interface PrivateProfileModalProps {
  isOpen: boolean;
  onClose: () => void;
  userName: string;
  avatarSrc?: string;
  preview?: boolean;
}

export default function PrivateProfileModal({
  isOpen,
  onClose,
  userName,
  avatarSrc,
  preview = false,
}: PrivateProfileModalProps) {
  const [requestSent, setRequestSent] = useState(false);

  // Reset state when modal closes
  useEffect(() => {
    if (!isOpen) {
      setRequestSent(false);
    }
  }, [isOpen]);

  // Auto-close after 2-3 seconds when request is sent
  useEffect(() => {
    if (requestSent) {
      const timer = setTimeout(() => {
        onClose();
      }, 2500); // 2.5 seconds
      return () => clearTimeout(timer);
    }
  }, [requestSent, onClose]);

  const handleSendRequest = () => {
    setRequestSent(true);
  };

  if (!isOpen && !preview) return null;

  const modalContent = (
    <div
      className="
        relative
        w-[349px]
        h-[427px]
        p-8
        flex
        flex-col
        justify-center
        items-center
        bg-parea-white
        border
        border-parea-black
      "
    >
      {/* Close Button */}
      <IconButton
        variant="close"
        onClick={onClose}
        aria-label="Close modal"
        className="
          absolute
          right-4
          top-4
          flex
          p-3
          justify-center
          items-center
        "
      />

      {/* Inner Content Container */}
      <div
        className="
          flex
          h-[362px]
          flex-col
          justify-center
          items-center
          self-stretch
          gap-4
        "
      >
        {/* Avatar */}
        <div
          className="
            w-40
            h-40
            aspect-square
            rounded-[160px]
            border
            border-parea-black
          "
          style={{
            background: avatarSrc
              ? `url(${avatarSrc}) lightgray 50% / cover no-repeat`
              : 'lightgray',
            backgroundSize: 'cover',
            backgroundPosition: 'center',
          }}
        />

        {/* User Info Container */}
        <div
          className="
            flex
            flex-col
            justify-center
            items-center
            gap-4
            self-stretch
          "
        >
          {/* Username and Message Container */}
          <div
            className="
              flex
              flex-col
              items-center
              gap-4
              flex-1
              self-stretch
            "
          >
            {/* Username */}
            <h2
              className="
                self-stretch
                text-center
                text-parea-black
                label-lg
                uppercase
              "
            >
              {userName.toUpperCase()}
            </h2>

            {/* Text Message */}
            {!requestSent && (
              <p
                className="
                  text-parea-black
                  text-center
                  self-stretch
                  text-regular
                "
              >
                This account is private. Send a follow request to get access.
              </p>
            )}

            {/* Request Sent Message */}
            {requestSent && (
              <p
                className="
                  text-parea-black
                  text-center
                  self-stretch
                  text-regular
                  underline
                  label-lg
                "
              >
                REQUEST SENT!
              </p>
            )}
          </div>

          {/* Send Follow Request Button */}
          {!requestSent && (
            <Button
              variant="primary"
              size="sm"
              onClick={handleSendRequest}
            >
              SEND FOLLOW REQUEST
            </Button>
          )}
        </div>
      </div>
    </div>
  );

  if (preview) {
    return modalContent;
  }

  return (
    <div
      className="
        fixed
        inset-0
        z-50
        flex
        items-center
        justify-center
      "
      onClick={onClose}
    >
      <div
        className="
          absolute
          inset-0
          bg-black/50
        "
      />
      <div onClick={(e) => e.stopPropagation()}>
        {modalContent}
      </div>
    </div>
  );
}
