'use client';

import { useState, useRef, useEffect } from 'react';
import IconButton from './IconButtons';
import Image from 'next/image';
import { getConversation, sendMessage, type ChatMessage } from '@/lib/api';

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
  /** User ID of the other user */
  userId: string;
  /** Avatar URL of the other user */
  userAvatar?: string;
  /** Avatar URL of the current user */
  selfAvatar?: string;
}

export default function ChatModal({
  isOpen,
  onClose,
  preview = false,
  userName,
  userId,
  userAvatar = '/test-avatar.png',
  selfAvatar = '/test-avatar.png',
}: ChatModalProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const currentUserIdRef = useRef<string | null>(null);

  // Load messages when modal opens or userId changes
  useEffect(() => {
    if (!isOpen || !userId) return;

    const loadMessages = async () => {
      setIsLoading(true);
      try {
        const data = await getConversation(userId, 50, 0);
        const formattedMessages: Message[] = data.messages.map((msg: ChatMessage) => ({
          id: msg.message_id,
          text: msg.content,
          sender: msg.sender_id === userId ? 'other' : 'self',
          avatarUrl: msg.sender_id === userId ? userAvatar : selfAvatar,
        })).reverse(); // Reverse to show oldest first
        
        setMessages(formattedMessages);
      } catch (error) {
        console.error('Failed to load messages:', error);
      } finally {
        setIsLoading(false);
      }
    };

    loadMessages();
  }, [isOpen, userId, userAvatar, selfAvatar]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  if (!isOpen && !preview) return null;

  const handleSend = async () => {
    const trimmedValue = inputValue.trim();
    if (!trimmedValue || isSending) return;

    setIsSending(true);
    const tempId = Date.now().toString();
    
    // Optimistically add message to UI
    const optimisticMessage: Message = {
      id: tempId,
      text: trimmedValue,
      sender: 'self',
      avatarUrl: selfAvatar,
    };
    
    setMessages((prev) => [...prev, optimisticMessage]);
    setInputValue('');

    try {
      const response = await sendMessage(userId, trimmedValue);
      
      // Replace temp message with real message from backend
      setMessages((prev) => 
        prev.map((msg) => 
          msg.id === tempId 
            ? { ...msg, id: response.message.message_id } 
            : msg
        )
      );
    } catch (error) {
      console.error('Failed to send message:', error);
      // Remove failed message
      setMessages((prev) => prev.filter((msg) => msg.id !== tempId));
      // Restore input value
      setInputValue(trimmedValue);
    } finally {
      setIsSending(false);
    }
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
    <div className="relative w-full max-w-145 bg-white border border-parea-black shadow-[8px_8px_0_0_#000] flex flex-col h-[600px]">
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
                <div className="w-[49px] h-[49px] rounded-full overflow-hidden relative shrink-0">
                  <Image src={avatar} alt="User" fill className="object-cover" />
                </div>
              )}

              {/* Bubble */}
              <div
                className={`relative w-[300px] p-4 border border-[#ccc] ${
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
                    className="absolute top-[13px] -right-[10px] w-0 h-0"
                    style={{
                      borderTop: '6px solid transparent',
                      borderBottom: '6px solid transparent',
                      borderLeft: '10px solid #ccc',
                    }}
                  >
                    <div
                      className="absolute top-[-5px] left-[-11px] w-0 h-0"
                      style={{
                        borderTop: '5px solid transparent',
                        borderBottom: '5px solid transparent',
                        borderLeft: '9px solid rgba(221,255,48,0.5)',
                      }}
                    />
                  </div>
                ) : (
                  <div
                    className="absolute top-[13px] -left-[10px] w-0 h-0"
                    style={{
                      borderTop: '6px solid transparent',
                      borderBottom: '6px solid transparent',
                      borderRight: '10px solid #ccc',
                    }}
                  >
                    <div
                      className="absolute top-[-5px] -right-[-2px] w-0 h-0"
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
                <div className="w-[49px] h-[49px] rounded-full overflow-hidden relative shrink-0">
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
