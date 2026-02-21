'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
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

function formatTime(value: string): string {
  if (!value) return '';
  try {
    return new Date(value).toLocaleString();
  } catch {
    return value;
  }
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
  const [inputValue, setInputValue] = useState('');

  const [isLoadingList, setIsLoadingList] = useState(true);
  const [isLoadingMessages, setIsLoadingMessages] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [chatError, setChatError] = useState<string | null>(null);

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
    setInputValue('');
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
    setInputValue('');
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

  const handleSend = useCallback(async () => {
    const content = inputValue.trim();
    if (!content || isSending) return;

    setIsSending(true);
    setChatError(null);

    try {
      if (activeTab === 'DIRECT CHATS' && selectedDirectUser) {
        const message = await sendChatMessage(selectedDirectUser.id, content);
        mergeDirectMessage(message);
        touchConversation(message);
      } else if (activeTab === 'GROUP CHATS' && selectedGroup) {
        const message = await sendGroupChatMessage(selectedGroup.id, content);
        mergeGroupMessage(message);
      }
      setInputValue('');
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to send message';
      setChatError(message);
    } finally {
      setIsSending(false);
    }
  }, [
    activeTab,
    inputValue,
    isSending,
    mergeDirectMessage,
    mergeGroupMessage,
    selectedDirectUser,
    selectedGroup,
    touchConversation,
  ]);

  const onInputKeyDown = useCallback((event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      void handleSend();
    }
  }, [handleSend]);

  return (
    <>
      <div className="flex pt-5 px-2 pb-3 justify-center items-center self-stretch">
        <Tabs
          tabs={['DIRECT CHATS', 'GROUP CHATS']}
          defaultTab="DIRECT CHATS"
          onTabChange={(tab) => {
            const nextTab = (tab === 'GROUP CHATS' ? 'GROUP CHATS' : 'DIRECT CHATS') as ChatTab;
            setActiveTab(nextTab);
            setInputValue('');
            setChatError(null);
          }}
          className="w-full justify-center"
        />
      </div>

      {chatError && (
        <div className="mx-4 mb-3 rounded border border-parea-black bg-parea-yellow/30 px-3 py-2 text-xs text-parea-black">
          {chatError}
        </div>
      )}

      {activeTab === 'DIRECT CHATS' ? (
        selectedDirectUser ? (
          <div className="flex flex-1 flex-col self-stretch overflow-hidden">
            <div className="flex items-center gap-2 border-b border-parea-black px-4 py-3">
              <button
                type="button"
                className="rounded border border-parea-black bg-parea-white px-2 py-1 text-xs font-medium uppercase text-parea-black hover:opacity-90"
                onClick={() => {
                  setSelectedDirectUser(null);
                  setDirectMessages([]);
                  setInputValue('');
                }}
              >
                Back
              </button>
              <span className="text-sm font-medium uppercase text-parea-black">
                {selectedDirectUser.nickname}
              </span>
            </div>

            <div ref={messageContainerRef} className="flex-1 overflow-y-auto px-4 py-3">
              {isLoadingMessages ? (
                <p className="text-sm text-parea-black">Loading conversation...</p>
              ) : directMessages.length === 0 ? (
                <p className="text-sm text-parea-black">No messages yet.</p>
              ) : (
                <div className="flex flex-col gap-2">
                  {directMessages.map((message) => {
                    const isMine = message.sender_id === currentUserId;
                    return (
                      <div
                        key={message.message_id}
                        className={`max-w-[85%] rounded border border-parea-black px-3 py-2 text-sm ${
                          isMine ? 'ml-auto bg-parea-yellow/40' : 'mr-auto bg-parea-white'
                        }`}
                      >
                        {!isMine && (
                          <p className="text-[11px] uppercase text-parea-black/70">{message.sender_name}</p>
                        )}
                        <p className="whitespace-pre-wrap break-words text-parea-black">{message.content}</p>
                        <p className="mt-1 text-[10px] text-parea-black/60">{formatTime(message.created_at)}</p>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>

            <div className="border-t border-parea-black px-4 py-3">
              <div className="flex gap-2">
                <input
                  value={inputValue}
                  onChange={(event) => setInputValue(event.target.value)}
                  onKeyDown={onInputKeyDown}
                  placeholder="Type a message..."
                  className="w-full rounded border border-parea-black bg-parea-white px-3 py-2 text-sm text-parea-black focus:outline-none"
                />
                <button
                  type="button"
                  onClick={() => {
                    void handleSend();
                  }}
                  disabled={isSending || inputValue.trim().length === 0}
                  className="rounded border border-parea-black bg-parea-yellow px-3 py-2 text-xs font-medium uppercase text-parea-black hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  Send
                </button>
              </div>
            </div>
          </div>
        ) : (
          <div className="flex flex-1 flex-col self-stretch overflow-y-auto px-4 pb-4">
            {isLoadingList ? (
              <p className="text-sm text-parea-black">Loading chats...</p>
            ) : (
              <>
                <p className="mb-2 text-xs font-semibold uppercase text-parea-black/70">Conversations</p>
                {conversations.length === 0 ? (
                  <p className="mb-4 text-sm text-parea-black">No conversations yet.</p>
                ) : (
                  <div className="mb-4 flex flex-col gap-2">
                    {conversations.map((conversation) => (
                      <button
                        key={conversation.user_id}
                        type="button"
                        onClick={() => {
                          void openDirectConversation(conversation.user_id, conversation.nickname || 'User');
                        }}
                        className="flex items-center justify-between rounded border border-parea-black bg-parea-white px-3 py-2 text-left hover:bg-parea-yellow/20"
                      >
                        <div className="min-w-0">
                          <p className="truncate text-xs font-semibold uppercase text-parea-black">
                            {conversation.nickname || 'User'}
                          </p>
                          <p className="truncate text-[11px] text-parea-black/70">{conversation.last_message}</p>
                        </div>
                        {conversation.unread_count > 0 && (
                          <span className="ml-2 rounded-full border border-parea-black bg-parea-yellow px-2 py-0.5 text-[10px] font-semibold text-parea-black">
                            {conversation.unread_count}
                          </span>
                        )}
                      </button>
                    ))}
                  </div>
                )}

                <p className="mb-2 text-xs font-semibold uppercase text-parea-black/70">Start New Chat</p>
                {newChatCandidates.length === 0 ? (
                  <p className="text-sm text-parea-black">No available users.</p>
                ) : (
                  <div className="flex flex-col gap-2">
                    {newChatCandidates.map((user) => (
                      <button
                        key={user.id}
                        type="button"
                        onClick={() => {
                          void openDirectConversation(user.id, user.nickname || 'User');
                        }}
                        className="rounded border border-parea-black bg-parea-white px-3 py-2 text-left text-xs font-semibold uppercase text-parea-black hover:bg-parea-yellow/20"
                      >
                        {user.nickname || `${user.first_name} ${user.last_name}`.trim() || 'User'}
                      </button>
                    ))}
                  </div>
                )}
              </>
            )}
          </div>
        )
      ) : (
        selectedGroup ? (
          <div className="flex flex-1 flex-col self-stretch overflow-hidden">
            <div className="flex items-center gap-2 border-b border-parea-black px-4 py-3">
              <button
                type="button"
                className="rounded border border-parea-black bg-parea-white px-2 py-1 text-xs font-medium uppercase text-parea-black hover:opacity-90"
                onClick={() => {
                  setSelectedGroup(null);
                  setGroupMessages([]);
                  setInputValue('');
                }}
              >
                Back
              </button>
              <span className="text-sm font-medium uppercase text-parea-black">{selectedGroup.title}</span>
            </div>

            <div ref={messageContainerRef} className="flex-1 overflow-y-auto px-4 py-3">
              {isLoadingMessages ? (
                <p className="text-sm text-parea-black">Loading group room...</p>
              ) : groupMessages.length === 0 ? (
                <p className="text-sm text-parea-black">No group messages yet.</p>
              ) : (
                <div className="flex flex-col gap-2">
                  {groupMessages.map((message) => {
                    const isMine = message.sender_id === currentUserId;
                    return (
                      <div
                        key={message.message_id}
                        className={`max-w-[85%] rounded border border-parea-black px-3 py-2 text-sm ${
                          isMine ? 'ml-auto bg-parea-yellow/40' : 'mr-auto bg-parea-white'
                        }`}
                      >
                        {!isMine && (
                          <p className="text-[11px] uppercase text-parea-black/70">{message.sender_name}</p>
                        )}
                        <p className="whitespace-pre-wrap break-words text-parea-black">{message.content}</p>
                        <p className="mt-1 text-[10px] text-parea-black/60">{formatTime(message.created_at)}</p>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>

            <div className="border-t border-parea-black px-4 py-3">
              <div className="flex gap-2">
                <input
                  value={inputValue}
                  onChange={(event) => setInputValue(event.target.value)}
                  onKeyDown={onInputKeyDown}
                  placeholder="Type a message..."
                  className="w-full rounded border border-parea-black bg-parea-white px-3 py-2 text-sm text-parea-black focus:outline-none"
                />
                <button
                  type="button"
                  onClick={() => {
                    void handleSend();
                  }}
                  disabled={isSending || inputValue.trim().length === 0}
                  className="rounded border border-parea-black bg-parea-yellow px-3 py-2 text-xs font-medium uppercase text-parea-black hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
                >
                  Send
                </button>
              </div>
            </div>
          </div>
        ) : (
          <div className="flex flex-1 flex-col self-stretch overflow-y-auto px-4 pb-4">
            {isLoadingList ? (
              <p className="text-sm text-parea-black">Loading groups...</p>
            ) : myGroups.length === 0 ? (
              <p className="text-sm text-parea-black">You are not a member of any groups yet.</p>
            ) : (
              <div className="flex flex-col gap-2">
                {myGroups.map((group) => (
                  <button
                    key={group.id}
                    type="button"
                    onClick={() => {
                      void openGroupConversation(group);
                    }}
                    className="rounded border border-parea-black bg-parea-white px-3 py-2 text-left text-xs font-semibold uppercase text-parea-black hover:bg-parea-yellow/20"
                  >
                    {group.title}
                  </button>
                ))}
              </div>
            )}
          </div>
        )
      )}
    </>
  );
}
