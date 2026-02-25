'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import Image from 'next/image';
import ChatModal from './ChatModal';
import Tabs from './Tabs';
import {
  getChatConversation,
  getChatConversations,
  getChatUsersForChat,
  getGroupChatMessages,
  getMyGroups,
  getProfile,
  sendChatMessage,
  sendGroupChatMessage,
  type ChatConversation,
  type DirectChatMessage,
  type ForumUser,
  type GroupChatMessage,
} from '@/lib/api';

type ChatTab = 'DIRECT CHATS' | 'GROUP CHATS';

type WebSocketEnvelope = {
  type?: string;
  data?: unknown;
};

type BasicGroup = {
  id: string;
  title: string;
};

function buildWebSocketURL(): string {
  const customURL = process.env.NEXT_PUBLIC_WS_URL;
  if (customURL) {
    return customURL;
  }

  const apiURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
  try {
    const url = new URL(apiURL);
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
    url.pathname = '/ws';
    url.search = '';
    return url.toString();
  } catch {
    return 'ws://localhost:8080/ws';
  }
}

function sortByCreatedAtAsc<T extends { created_at: string }>(items: T[]): T[] {
  return [...items].sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
}

function dedupeById<T extends { message_id: string }>(items: T[]): T[] {
  const byId = new Map<string, T>();
  items.forEach((item) => {
    byId.set(item.message_id, item);
  });
  return Array.from(byId.values());
}

function normalizeDirectMessage(data: unknown): DirectChatMessage | null {
  if (!data || typeof data !== 'object') {
    return null;
  }

  const payload = data as Partial<DirectChatMessage>;
  if (
    typeof payload.message_id !== 'string' ||
    typeof payload.sender_id !== 'string' ||
    typeof payload.sender_name !== 'string' ||
    typeof payload.receiver_id !== 'string' ||
    typeof payload.content !== 'string' ||
    typeof payload.created_at !== 'string' ||
    typeof payload.is_read !== 'boolean'
  ) {
    return null;
  }

  return payload as DirectChatMessage;
}

function normalizeGroupMessage(data: unknown): GroupChatMessage | null {
  if (!data || typeof data !== 'object') {
    return null;
  }

  const payload = data as Partial<GroupChatMessage>;
  if (
    typeof payload.message_id !== 'string' ||
    typeof payload.group_id !== 'string' ||
    typeof payload.sender_id !== 'string' ||
    typeof payload.sender_name !== 'string' ||
    typeof payload.content !== 'string' ||
    typeof payload.created_at !== 'string'
  ) {
    return null;
  }

  return payload as GroupChatMessage;
}

export default function ChatDrawer() {
  const [activeTab, setActiveTab] = useState<ChatTab>('DIRECT CHATS');
  const [currentUserId, setCurrentUserId] = useState('');

  const [conversations, setConversations] = useState<ChatConversation[]>([]);
  const [chatUsers, setChatUsers] = useState<ForumUser[]>([]);
  const [myGroups, setMyGroups] = useState<BasicGroup[]>([]);

  const [selectedDirectUser, setSelectedDirectUser] = useState<{ id: string; nickname: string } | null>(null);
  const [selectedGroup, setSelectedGroup] = useState<BasicGroup | null>(null);

  const [directMessages, setDirectMessages] = useState<DirectChatMessage[]>([]);
  const [groupMessages, setGroupMessages] = useState<GroupChatMessage[]>([]);
  const [isLoadingList, setIsLoadingList] = useState(true);
  const [, setIsLoadingMessages] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [, setChatError] = useState<string | null>(null);
  const [chatModalOpen, setChatModalOpen] = useState(false);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const messageContainerRef = useRef<HTMLDivElement | null>(null);
  const currentUserIdRef = useRef<string>('');
  const selectedDirectUserRef = useRef<{ id: string; nickname: string } | null>(null);
  const selectedGroupRef = useRef<BasicGroup | null>(null);

  const conversationUserIDs = useMemo(() => {
    return new Set(conversations.map((conversation) => conversation.user_id));
  }, [conversations]);

  const newChatCandidates = useMemo(() => {
    return chatUsers.filter((user) => !conversationUserIDs.has(user.id));
  }, [chatUsers, conversationUserIDs]);

  useEffect(() => {
    currentUserIdRef.current = currentUserId;
  }, [currentUserId]);

  useEffect(() => {
    selectedDirectUserRef.current = selectedDirectUser;
  }, [selectedDirectUser]);

  useEffect(() => {
    selectedGroupRef.current = selectedGroup;
  }, [selectedGroup]);

  const mergeDirectMessage = useCallback((message: DirectChatMessage) => {
    setDirectMessages((prev) => {
      return sortByCreatedAtAsc(dedupeById([...prev, message]));
    });
  }, []);

  const mergeGroupMessage = useCallback((message: GroupChatMessage) => {
    setGroupMessages((prev) => {
      return sortByCreatedAtAsc(dedupeById([...prev, message]));
    });
  }, []);

  const touchConversation = useCallback((message: DirectChatMessage) => {
    const selfID = currentUserIdRef.current;
    if (!selfID) return;

    const otherUserID = message.sender_id === selfID ? message.receiver_id : message.sender_id;
    const currentSelection = selectedDirectUserRef.current;
    const otherNickname = message.sender_id === selfID ? currentSelection?.nickname || 'User' : message.sender_name;

    setConversations((prev) => {
      const updated = [...prev];
      const existingIndex = updated.findIndex((item) => item.user_id === otherUserID);
      const isActiveConversation = currentSelection?.id === otherUserID;

      if (existingIndex >= 0) {
        const existing = updated[existingIndex];
        const unreadCount = message.sender_id === selfID || isActiveConversation ? 0 : existing.unread_count + 1;
        updated[existingIndex] = {
          ...existing,
          nickname: existing.nickname || otherNickname,
          last_message: message.content,
          last_message_time: message.created_at,
          unread_count: unreadCount,
        };
      } else {
        updated.push({
          user_id: otherUserID,
          nickname: otherNickname,
          last_message: message.content,
          last_message_time: message.created_at,
          unread_count: message.sender_id === selfID || isActiveConversation ? 0 : 1,
          is_online: false,
        });
      }

      return updated.sort(
        (a, b) =>
          new Date(b.last_message_time || 0).getTime() -
          new Date(a.last_message_time || 0).getTime(),
      );
    });
  }, []);

  const loadBaseData = useCallback(async () => {
    setIsLoadingList(true);
    setChatError(null);
    try {
      const [profile, directConversations, users, groups] = await Promise.all([
        getProfile(),
        getChatConversations(),
        getChatUsersForChat(true),
        getMyGroups(),
      ]);

      setCurrentUserId(typeof profile?.id === 'string' ? profile.id : '');
      setConversations(directConversations);
      setChatUsers(users);
      setMyGroups(
        (Array.isArray(groups) ? groups : [])
          .filter((group) => typeof group?.id === 'string')
          .map((group) => ({
            id: group.id as string,
            title: typeof group?.title === 'string' ? group.title : 'Group',
          })),
      );
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load chat data';
      setChatError(message);
      setConversations([]);
      setChatUsers([]);
      setMyGroups([]);
    } finally {
      setIsLoadingList(false);
    }
  }, []);

  useEffect(() => {
    void loadBaseData();
  }, [loadBaseData]);

  useEffect(() => {
    if (!currentUserId) {
      return;
    }

    let isUnmounted = false;
    const wsURL = buildWebSocketURL();

    const processWebSocketMessage = (raw: string) => {
      const payloads = raw
        .split('\n')
        .map((line) => line.trim())
        .filter((line) => line.length > 0);

      for (const payload of payloads) {
        try {
          const parsed = JSON.parse(payload) as WebSocketEnvelope;

          if (parsed.type === 'chat') {
            const directMessage = normalizeDirectMessage(parsed.data);
            if (!directMessage) {
              continue;
            }
            const selfID = currentUserIdRef.current;
            if (selfID && directMessage.sender_id !== selfID && directMessage.receiver_id !== selfID) {
              continue;
            }

            touchConversation(directMessage);
            const activeDirect = selectedDirectUserRef.current;
            if (
              activeDirect &&
              (activeDirect.id === directMessage.sender_id || activeDirect.id === directMessage.receiver_id)
            ) {
              mergeDirectMessage(directMessage);
            }
            continue;
          }

          if (parsed.type === 'group_chat') {
            const groupMessage = normalizeGroupMessage(parsed.data);
            if (!groupMessage) {
              continue;
            }
            const activeGroup = selectedGroupRef.current;
            if (activeGroup?.id === groupMessage.group_id) {
              mergeGroupMessage(groupMessage);
            }
          }
        } catch (error) {
          console.error('Failed to parse websocket payload:', error);
        }
      }
    };

    const connect = () => {
      if (isUnmounted) {
        return;
      }

      try {
        const socket = new WebSocket(wsURL);
        wsRef.current = socket;

        socket.onmessage = (event: MessageEvent) => {
          if (typeof event.data === 'string') {
            processWebSocketMessage(event.data);
            return;
          }

          if (event.data instanceof Blob) {
            event.data
              .text()
              .then(processWebSocketMessage)
              .catch((error) => {
                console.error('Failed to read websocket blob payload:', error);
              });
          }
        };

        socket.onerror = () => {
          socket.close();
        };

        socket.onclose = () => {
          if (isUnmounted) {
            return;
          }
          reconnectTimerRef.current = setTimeout(connect, 3000);
        };
      } catch (error) {
        console.error('Failed to connect websocket for chat:', error);
        reconnectTimerRef.current = setTimeout(connect, 3000);
      }
    };

    connect();

    return () => {
      isUnmounted = true;

      if (reconnectTimerRef.current) {
        clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }

      if (wsRef.current) {
        wsRef.current.onclose = null;
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [currentUserId, mergeDirectMessage, mergeGroupMessage, touchConversation]);

  useEffect(() => {
    if (!messageContainerRef.current) return;
    messageContainerRef.current.scrollTop = messageContainerRef.current.scrollHeight;
  }, [directMessages, groupMessages, selectedDirectUser, selectedGroup]);

  const openDirectConversation = useCallback(async (userID: string, nickname: string) => {
    setIsLoadingMessages(true);
    setChatError(null);
    setSelectedGroup(null);
    setSelectedDirectUser({ id: userID, nickname });
    try {
      const messages = await getChatConversation(userID, 100, 0);
      setDirectMessages(sortByCreatedAtAsc(messages));
      setConversations((prev) => {
        return prev.map((conversation) =>
          conversation.user_id === userID ? { ...conversation, unread_count: 0 } : conversation,
        );
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load conversation';
      setChatError(message);
      setDirectMessages([]);
    } finally {
      setIsLoadingMessages(false);
    }
  }, []);

  const openGroupConversation = useCallback(async (group: BasicGroup) => {
    setIsLoadingMessages(true);
    setChatError(null);
    setSelectedDirectUser(null);
    setSelectedGroup(group);
    try {
      const messages = await getGroupChatMessages(group.id, 200, 0);
      setGroupMessages(sortByCreatedAtAsc(messages));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load group chat';
      setChatError(message);
      setGroupMessages([]);
    } finally {
      setIsLoadingMessages(false);
    }
  }, []);

  const sendDirectMessage = useCallback(async (content: string) => {
    if (!content.trim() || isSending || !selectedDirectUser) return;
    setIsSending(true);
    setChatError(null);
    try {
      const message = await sendChatMessage(selectedDirectUser.id, content);
      mergeDirectMessage(message);
      touchConversation(message);
    } catch (error) {
      setChatError(error instanceof Error ? error.message : 'Failed to send message');
    } finally {
      setIsSending(false);
    }
  }, [isSending, mergeDirectMessage, selectedDirectUser, touchConversation]);


const chatModalMessages = useMemo(() => {
    return directMessages.map((msg) => ({
      id: msg.message_id,
      text: msg.content,
      sender: (msg.sender_id === currentUserId ? 'self' : 'other') as 'self' | 'other',
    }));
  }, [directMessages, currentUserId]);

  const groupChatModalMessages = useMemo(() => {
    return groupMessages.map((msg) => ({
      id: msg.message_id,
      text: msg.content,
      sender: (msg.sender_id === currentUserId ? 'self' : 'other') as 'self' | 'other',
    }));
  }, [groupMessages, currentUserId]);

  const sendGroupMessage = useCallback(async (content: string) => {
    if (!content.trim() || !selectedGroup) return;
    try {
      const message = await sendGroupChatMessage(selectedGroup.id, content);
      mergeGroupMessage(message);
    } catch (error) {
      setChatError(error instanceof Error ? error.message : 'Failed to send message');
    }
  }, [selectedGroup, mergeGroupMessage]);

  const handleCloseChatModal = useCallback(() => {
    setChatModalOpen(false);
    setSelectedDirectUser(null);
    setDirectMessages([]);
  }, []);

  const handleOpenConversation = useCallback((userId: string, nickname: string) => {
    void openDirectConversation(userId, nickname);
    setChatModalOpen(true);
  }, [openDirectConversation]);

  return (
    <>
      {/* Tab toggle */}
      <div className="flex items-center justify-center pt-5 pb-3 px-2 w-full shrink-0">
        <Tabs
          tabs={['DIRECT CHATS', 'GROUP CHATS']}
          defaultTab="DIRECT CHATS"
          onTabChange={(tab) => {
            setActiveTab(tab as ChatTab);
            setChatError(null);
          }}
          className="w-full justify-center"
        />
      </div>

      {/* List */}
      <div className="flex flex-col items-start pt-2 w-full overflow-y-auto flex-1">
        {isLoadingList ? (
          <p className="px-4 py-3 text-sm text-parea-black">Loading...</p>
        ) : activeTab === 'DIRECT CHATS' ? (
          <>
            {conversations.length === 0 && newChatCandidates.length === 0 && (
              <p className="px-4 py-3 font-mono text-sm uppercase tracking-[-0.16px] text-parea-black/50">
                No chats yet
              </p>
            )}
            {conversations.map((conversation) => (
              <button
                key={conversation.user_id}
                type="button"
                onClick={() => handleOpenConversation(conversation.user_id, conversation.nickname || 'User')}
                className="flex h-14 items-center justify-between px-4 py-2 w-full hover:bg-parea-grey/30 transition-colors"
              >
                <div className="flex flex-1 gap-2 items-center min-w-0">
                  <div className="relative shrink-0 size-10 rounded-full overflow-hidden bg-parea-grey">
                    <Image src="/user-avatar-default.png" alt="" fill className="object-cover" />
                  </div>
                  <p className="font-mono text-base font-medium uppercase tracking-[-0.16px] text-parea-black truncate">
                    {conversation.nickname || 'User'}
                  </p>
                </div>
                {conversation.unread_count > 0 && (
                  <div className="bg-parea-yellow border border-parea-black flex items-center justify-center px-2 rounded-full shrink-0 ml-2">
                    <span className="font-mono text-base font-medium uppercase tracking-[-0.16px] leading-relaxed">
                      {conversation.unread_count}
                    </span>
                  </div>
                )}
              </button>
            ))}
            {newChatCandidates.map((user) => (
              <button
                key={user.id}
                type="button"
                onClick={() => handleOpenConversation(user.id, user.nickname || `${user.first_name} ${user.last_name}`.trim() || 'User')}
                className="flex h-14 items-center px-4 py-2 w-full hover:bg-parea-grey/30 transition-colors"
              >
                <div className="flex flex-1 gap-2 items-center min-w-0">
                  <div className="relative shrink-0 size-10 rounded-full overflow-hidden bg-parea-grey">
                    <Image src="/user-avatar-default.png" alt="" fill className="object-cover" />
                  </div>
                  <p className="font-mono text-base font-medium uppercase tracking-[-0.16px] text-parea-black truncate">
                    {user.nickname || `${user.first_name} ${user.last_name}`.trim() || 'User'}
                  </p>
                </div>
              </button>
            ))}
          </>
        ) : (
          <>
            {myGroups.length === 0 && (
              <p className="px-4 py-3 font-mono text-sm uppercase tracking-[-0.16px] text-parea-black/50">
                No group chats yet
              </p>
            )}
            {myGroups.map((group) => (
              <button
                key={group.id}
                type="button"
                onClick={() => { void openGroupConversation(group); setChatModalOpen(true); }}
                className="flex h-14 items-center px-4 py-2 w-full hover:bg-parea-grey/30 transition-colors"
              >
                <div className="flex flex-1 gap-2 items-center min-w-0">
                  <div className="relative shrink-0 size-10 rounded-full overflow-hidden bg-parea-grey">
                    <Image src="/user-avatar-default.png" alt="" fill className="object-cover" />
                  </div>
                  <p className="font-mono text-base font-medium uppercase tracking-[-0.16px] text-parea-black truncate">
                    {group.title}
                  </p>
                </div>
              </button>
            ))}
          </>
        )}
      </div>

      {/* Chat modal */}
      {selectedDirectUser && (
        <ChatModal
          isOpen={chatModalOpen}
          onClose={handleCloseChatModal}
          userName={selectedDirectUser.nickname}
          controlledMessages={chatModalMessages}
          onSendMessage={(text) => { void sendDirectMessage(text); }}
        />
      )}
      {selectedGroup && (
        <ChatModal
          isOpen={chatModalOpen}
          onClose={() => { setChatModalOpen(false); setSelectedGroup(null); setGroupMessages([]); }}
          userName={selectedGroup.title}
          controlledMessages={groupChatModalMessages}
          onSendMessage={(text) => { void sendGroupMessage(text); }}
        />
      )}
    </>
  );
}
