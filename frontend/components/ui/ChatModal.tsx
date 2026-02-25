'use client';

import { useState, useRef, useEffect } from 'react';
import { createPortal } from 'react-dom';
import IconButton from './IconButtons';
import Image from 'next/image';

interface Message {
  id: string;
  text: string;
  sender: 'self' | 'other';
  avatarUrl?: string;
}

interface ChatModalProps {
  isOpen: boolean;
  onClose: () => void;
  preview?: boolean;
  /** Display name of the other user */
  userName: string;
  /** Avatar URL of the other user */
  userAvatar?: string;
  /** Avatar URL of the current user */
  selfAvatar?: string;
  /** Initial messages for demo/preview */
  initialMessages?: Message[];
  /** Controlled messages — when provided, these are displayed instead of internal state */
  controlledMessages?: Message[];
  /** Called when user sends a message — when provided, real send is used instead of internal mock */
  onSendMessage?: (text: string) => void;
}

const mockMessages: Message[] = [
  { id: '1', text: 'Hey there!', sender: 'other', avatarUrl: '/test-avatar.png' },
  {
    id: '2',
    text: 'Wanted to thank you for help yesterday! You are so kind and I thought I should reach you out to tell how I am appreciated.',
    sender: 'other',
    avatarUrl: '/test-avatar.png',
  },
  { id: '3', text: 'Ohh come on! Anyone would the same!', sender: 'self', avatarUrl: '/test-avatar.png' },
];

export default function ChatModal({
  isOpen,
  onClose,
  preview = false,
  userName,
  userAvatar = '/test-avatar.png',
  selfAvatar = '/test-avatar.png',
  initialMessages = mockMessages,
  controlledMessages,
  onSendMessage,
}: ChatModalProps) {
  const [internalMessages, setInternalMessages] = useState<Message[]>(initialMessages);
  const [inputValue, setInputValue] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const messages = controlledMessages ?? internalMessages;

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  if (!isOpen && !preview) return null;

  const handleSend = () => {
    if (!inputValue.trim()) return;
    if (onSendMessage) {
      onSendMessage(inputValue.trim());
    } else {
      setInternalMessages((prev) => [
        ...prev,
        {
          id: Date.now().toString(),
          text: inputValue.trim(),
          sender: 'self',
          avatarUrl: selfAvatar,
        },
      ]);
    }
    setInputValue('');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleClose = () => {
    onClose();
  };

  const modalContent = (
    <div className="relative w-95 bg-white border border-parea-black shadow-[8px_8px_0_0_#000] flex flex-col h-120">
      {/* ===== HEADER ===== */}
      <div className="relative flex items-center justify-between px-8 h-17 border-b border-parea-black overflow-hidden bg-parea-white shrink-0">
        <Image
          src="/chat-header-pattern.svg"
          alt=""
          fill
          className="object-cover"
        />
        <h4 className="relative z-10">{userName}</h4>
        <IconButton
          variant="close"
          onClick={handleClose}
          aria-label="Close modal"
          className="relative z-10"
        />
      </div>

      {/* ===== MESSAGES AREA ===== */}
      <div className="flex-1 overflow-y-auto p-4 flex flex-col gap-4">
        {messages.map((msg) => {
          const isSelf = msg.sender === 'self';
          const avatar = msg.avatarUrl || (isSelf ? selfAvatar : userAvatar);

          return (
            <div
              key={msg.id}
              className={`flex gap-4 items-start ${isSelf ? 'justify-end' : ''}`}
            >
              {/* Other user avatar (left) */}
              {!isSelf && (
                <div className="w-12 h-12 rounded-full overflow-hidden relative shrink-0">
                  <Image src={avatar} alt="User" fill className="object-cover" />
                </div>
              )}

              {/* Bubble */}
              <div
                className={`relative w-75 p-2 border border-[#ccc] ${
                  isSelf ? 'bg-parea-yellow/50' : 'bg-white'
                }`}
              >
                <p
                  className="text-regular leading-relaxed text-black"
                  style={{ fontFamily: 'var(--font-inter), sans-serif' }}
                >
                  {msg.text}
                </p>

                {/* Sharktooth pointer */}
                {isSelf ? (
                  <div
                    className="absolute top-3 -right-3 w-0 h-0"
                    style={{
                      borderTop: '6px solid transparent',
                      borderBottom: '6px solid transparent',
                      borderLeft: '10px solid #ccc',
                    }}
                  >
                    <div
                      className="absolute -top-1 -left-3 w-0 h-0"
                      style={{
                        borderTop: '5px solid transparent',
                        borderBottom: '5px solid transparent',
                        borderLeft: '9px solid rgba(221,255,48,0.5)',
                      }}
                    />
                  </div>
                ) : (
                  <div
                    className="absolute top-3 -left-3 w-0 h-0"
                    style={{
                      borderTop: '6px solid transparent',
                      borderBottom: '6px solid transparent',
                      borderRight: '10px solid #ccc',
                    }}
                  >
                    <div
                      className="absolute -top-1 right-1 w-0 h-0"
                      style={{
                        borderTop: '5px solid transparent',
                        borderBottom: '5px solid transparent',
                        borderRight: '9px solid white',
                      }}
                    />
                  </div>
                )}
              </div>

              {/* Self avatar (right) */}
              {isSelf && (
                <div className="w-12 h-12 rounded-full overflow-hidden relative shrink-0">
                  <Image src={avatar} alt="You" fill className="object-cover" />
                </div>
              )}
            </div>
          );
        })}
        <div ref={messagesEndRef} />
      </div>

      {/* ===== INPUT BAR ===== */}
      <div className="shrink-0 flex items-center bg-parea-white border-t border-parea-black pl-4 pr-8 py-4 gap-4">
        <input
          type="text"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Type a message"
          className="flex-1 bg-transparent text-regular leading-relaxed text-black placeholder:text-black/60 focus:outline-none"
          style={{ fontFamily: 'var(--font-inter), sans-serif' }}
        />
        <IconButton
          variant="arrow-right"
          onClick={handleSend}
          aria-label="Send message"
        />
      </div>
    </div>
  );

  if (preview) {
    return modalContent;
  }

  if (typeof window === 'undefined') return null;

  return createPortal(
    <div className="fixed bottom-4 right-4 z-9999">
      {modalContent}
    </div>,
    document.body,
  );
}
