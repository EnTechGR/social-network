// components/ui/Sidebar.tsx
'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { MessageCircle, Bell, LogOut } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import SidebarToggle from './SidebarToggle';
import Tabs from './Tabs';
import {
  logout,
  getProfile,
  getAvatarUrl,
  getNotifications,
  hideNotification,
  acceptFollowRequest,
  declineFollowRequest,
  type NotificationItem,
} from '@/lib/api';
import { clearAuth } from '@/lib/auth';

type DrawerType = 'chat' | 'notifications' | null;
type WebSocketEnvelope = {
  type?: string;
  data?: unknown;
  timestamp?: string;
};

const DEFAULT_AVATAR = '/user-avatar-default.png';

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

function formatNotificationText(notification: NotificationItem): string {
  const actor = notification.nickname || 'Someone';

  switch (notification.type) {
    case 'follow_request':
      return `${actor} sent you a follow request`;
    case 'follow_accept':
      return `${actor} accepted your follow request`;
    case 'comment':
    case 'post_comment':
      return `${actor} commented on your post`;
    case 'edit_comment':
      return `${actor} edited a comment on your post`;
    case 'delete_comment':
      return `${actor} deleted a comment on your post`;
    case 'like':
      return `${actor} liked your content`;
    case 'dislike':
      return `${actor} disliked your content`;
    case 'love':
      return `${actor} loved your content`;
    case 'post_reaction':
      return `${actor} reacted to your post`;
    case 'comment_reaction':
      return `${actor} reacted to your comment`;
    case 'group_invite':
      return `${actor} invited you to a group`;
    case 'group_join_request':
      return `${actor} requested to join your group`;
    case 'group_join_decision':
      return `${actor} responded to your group request`;
    case 'group_event':
      return `${actor} created a group event`;
    default:
      return `${actor} sent you a notification`;
  }
}

export default function Sidebar() {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [activeDrawer, setActiveDrawer] = useState<DrawerType>(null);
  const [lastDrawer, setLastDrawer] = useState<'chat' | 'notifications'>('chat');
  const [avatarUrl, setAvatarUrl] = useState(DEFAULT_AVATAR);
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [isLoadingNotifications, setIsLoadingNotifications] = useState(false);
  const [notificationsError, setNotificationsError] = useState<string | null>(null);
  const [followRequestActionId, setFollowRequestActionId] = useState<string | null>(null);
  const [followRequestActionError, setFollowRequestActionError] = useState<string | null>(null);
  const sidebarRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const unreadCount = notifications.reduce((count, notification) => {
    return notification.read ? count : count + 1;
  }, 0);

  const loadNotifications = useCallback(async () => {
    setIsLoadingNotifications(true);
    setNotificationsError(null);
    try {
      const data = await getNotifications();
      setNotifications(data.notifications ?? []);
    } catch (error) {
      console.error('Failed to fetch notifications:', error);
      setNotificationsError('Failed to load notifications');
    } finally {
      setIsLoadingNotifications(false);
    }
  }, []);

  useEffect(() => {
    getProfile()
      .then((p) => {
        const path = p?.avatar?.file_path || p?.avatar?.thumbnail_path;
        setAvatarUrl(getAvatarUrl(path));
      })
      .catch(() => { });
  }, []);

  useEffect(() => {
    void loadNotifications();
  }, [loadNotifications]);

  useEffect(() => {
    let isUnmounted = false;
    const wsURL = buildWebSocketURL();

    const pushNotification = (incoming: NotificationItem) => {
      if (!incoming?.id) {
        return;
      }

      setNotifications((prev) => {
        if (prev.some((notification) => notification.id === incoming.id)) {
          return prev;
        }
        return [incoming, ...prev];
      });
    };

    const processWebSocketMessage = (raw: string) => {
      const payloads = raw
        .split('\n')
        .map((line) => line.trim())
        .filter((line) => line.length > 0);

      for (const payload of payloads) {
        try {
          const parsed = JSON.parse(payload) as WebSocketEnvelope;
          if (parsed.type !== 'notification' || !parsed.data) {
            continue;
          }
          pushNotification(parsed.data as NotificationItem);
        } catch (error) {
          console.error('Failed to parse websocket message:', error);
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
        console.error('Failed to connect websocket:', error);
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
  }, []);

  const handleToggle = () => {
    if (activeDrawer) {
      setActiveDrawer(null);
      return;
    }
    setIsOpen(!isOpen);
  };

  const handleDrawerOpen = (type: 'chat' | 'notifications') => {
    if (activeDrawer === type) {
      setActiveDrawer(null);
      return;
    }
    setIsOpen(false);
    setActiveDrawer(type);
    setLastDrawer(type);
  };

  const handleDrawerClose = () => {
    setActiveDrawer(null);
  };

  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error('Logout API call failed:', error);
    } finally {
      clearAuth();
      router.push('/login');
    }
  };

  const handleDeleteNotification = async (notificationId: string) => {
    try {
      await hideNotification(notificationId);
      setNotifications((prev) => prev.filter((notification) => notification.id !== notificationId));
    } catch (error) {
      console.error('Failed to delete notification:', error);
      void loadNotifications();
    }
  };

  const handleFollowRequestAction = async (
    notification: NotificationItem,
    action: 'accept' | 'decline',
  ) => {
    const followerId = notification.from_user_id;
    if (!followerId) {
      setFollowRequestActionError('Cannot process this follow request right now.');
      return;
    }

    setFollowRequestActionId(notification.id);
    setFollowRequestActionError(null);

    try {
      if (action === 'accept') {
        await acceptFollowRequest(followerId);
      } else {
        await declineFollowRequest(followerId);
      }

      await hideNotification(notification.id);
      setNotifications((prev) => prev.filter((item) => item.id !== notification.id));
    } catch (error) {
      console.error(`Failed to ${action} follow request:`, error);
      setFollowRequestActionError(`Failed to ${action} follow request`);
    } finally {
      setFollowRequestActionId(null);
    }
  };

  const handleDeleteAllNotifications = async () => {
    if (notifications.length === 0) {
      return;
    }

    const snapshot = notifications;
    setNotifications([]);

    try {
      await Promise.all(snapshot.map((notification) => hideNotification(notification.id)));
    } catch (error) {
      console.error('Failed to delete all notifications:', error);
      void loadNotifications();
    }
  };

  const handleMarkAllAsRead = () => {
    setNotifications((prev) => prev.map((notification) => ({ ...notification, read: true })));
  };

  useEffect(() => {
    if (!isOpen && !activeDrawer) return;
    function handleClickOutside(event: MouseEvent) {
      if (sidebarRef.current && !sidebarRef.current.contains(event.target as Node)) {
        setActiveDrawer(null);
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen, activeDrawer]);

  return (
    <div ref={sidebarRef}>
      {/* Main Sidebar — single layout, width transitions */}
      <aside
        className={`
          fixed left-0 top-0 h-screen z-50
          flex flex-col items-start
          bg-parea-black border-r border-parea-black
          overflow-hidden
          transition-[width] duration-300 ease-in-out
          ${isOpen ? 'w-79' : 'w-18'}
        `}
      >
        {/* Top section: toggle + logo */}
        <div className="flex items-center gap-2 shrink-0 self-stretch pt-6 pb-4 px-4">
          <div className="w-10 shrink-0 flex items-center justify-center">
            <SidebarToggle isOpen={isOpen || !!activeDrawer} onClick={handleToggle} />
          </div>
          <Image
            src="/logo-light.svg"
            alt="Parea Logo"
            width={80}
            height={32}
            className={`h-8 w-auto shrink-0 transition-opacity duration-300 ${isOpen ? 'opacity-100' : 'opacity-0'}`}
          />
        </div>

        {/* Menu items */}
        <div className="flex flex-col justify-between flex-1 self-stretch">
          {/* Top menu */}
          <div className="flex flex-col items-start self-stretch pt-2 px-4">
            {/* Chat */}
            <button
              onClick={() => handleDrawerOpen('chat')}
              className="flex items-center gap-2 py-2 self-stretch rounded-lg transition-colors group cursor-pointer"
            >
              <div className="w-10 shrink-0 flex items-center justify-center">
                <MessageCircle
                  className={`w-6 h-6 shrink-0 transition-colors ${activeDrawer === 'chat'
                    ? 'text-parea-yellow'
                    : 'text-parea-white group-hover:text-parea-yellow'
                    }`}
                />
              </div>
              <span className={`text-parea-white font-mono text-sm font-medium uppercase tracking-wide group-hover:text-parea-yellow transition-opacity duration-300 whitespace-nowrap ${isOpen ? 'opacity-100' : 'opacity-0'}`}>
                Chats
              </span>
            </button>

            {/* Notifications */}
            <button
              onClick={() => handleDrawerOpen('notifications')}
              className="flex items-center justify-between gap-2 py-2 self-stretch rounded-lg transition-colors group cursor-pointer"
            >
              <div className="flex items-center gap-2">
                <div className="w-10 shrink-0 flex items-center justify-center">
                  <Bell
                    className={`w-6 h-6 shrink-0 transition-colors ${activeDrawer === 'notifications'
                      ? 'text-parea-yellow'
                      : 'text-parea-white group-hover:text-parea-yellow'
                      }`}
                  />
                </div>
                <span className={`text-parea-white font-mono text-sm font-medium uppercase tracking-wide group-hover:text-parea-yellow transition-opacity duration-300 whitespace-nowrap ${isOpen ? 'opacity-100' : 'opacity-0'}`}>
                  Notifications
                </span>
              </div>
              <span className={`flex items-center justify-center min-w-6 h-6 px-2 rounded-full border border-parea-yellow text-parea-yellow font-mono text-xs font-medium shrink-0 transition-opacity duration-300 ${isOpen ? 'opacity-100' : 'opacity-0'}`}>
                {unreadCount}
              </span>
            </button>
          </div>

          {/* Bottom menu: profile + logout */}
          <div className="flex flex-col items-start gap-2 self-stretch pb-8 px-4">
            {/* Profile */}
            <Link
              href="/profile"
              className="flex items-center gap-2 py-2 self-stretch rounded-lg transition-colors group cursor-pointer"
            >
              <div className="w-10 h-10 rounded-full border-2 border-parea-yellow overflow-hidden relative shrink-0">
                <Image
                  src={avatarUrl}
                  alt="User avatar"
                  fill
                  className="object-cover"
                />
              </div>
              <span className={`text-parea-white font-mono text-sm font-medium uppercase tracking-wide group-hover:text-parea-yellow transition-opacity duration-300 whitespace-nowrap ${isOpen ? 'opacity-100' : 'opacity-0'}`}>
                Profile
              </span>
            </Link>

            {/* Logout */}
            <button
              onClick={handleLogout}
              className="flex items-center gap-2 py-2 self-stretch rounded-lg transition-colors group cursor-pointer"
            >
              <div className="w-10 shrink-0 flex items-center justify-center">
                <LogOut className="w-6 h-6 shrink-0 text-parea-white group-hover:text-parea-yellow transition-colors" />
              </div>
              <span className={`text-parea-white font-mono text-sm font-medium uppercase tracking-wide group-hover:text-parea-yellow transition-opacity duration-300 whitespace-nowrap ${isOpen ? 'opacity-100' : 'opacity-0'}`}>
                Logout
              </span>
            </button>
          </div>
        </div>
      </aside>

      {/* Drawer (Chat or Notifications) */}
      <aside
        className={`
          fixed left-18 top-0 h-screen w-79 z-40
          flex flex-col items-start
          bg-parea-white border-r border-parea-black
          transition-transform duration-300 ease-in-out
          ${activeDrawer ? 'translate-x-0' : '-translate-x-full'}
        `}
      >
          {(activeDrawer || lastDrawer) === 'chat' ? (
            <>
              {/* Tabs Section */}
              <div className="flex pt-5 px-2 pb-3 justify-center items-center self-stretch">
                <Tabs
                  tabs={['DIRECT CHATS', 'GROUP CHATS']}
                  defaultTab="DIRECT CHATS"
                  className="w-full justify-center"
                />
              </div>

              {/* Chat List */}
              <div className="flex flex-1 pt-2 flex-col items-start gap-4 self-stretch overflow-y-auto px-4">
                  {[
                    { name: 'JENNIFER WHITE', count: 2, avatar: '/test-avatar.png' },
                    { name: 'ALEX DONHAM', count: 1, avatar: '/test-avatar.png' },
                    { name: 'KAREN HILLS', count: null, avatar: '/test-avatar.png' },
                  ].map((profile, index) => (
                    <div
                      key={index}
                      className="flex h-14 py-2 px-4 justify-between items-center self-stretch rounded-lg hover:bg-parea-grey/30 transition-colors cursor-pointer"
                    >
                      <div className="flex items-center gap-3">
                        <div className="w-10 h-10 rounded-full overflow-hidden relative shrink-0">
                          <Image src={profile.avatar} alt={profile.name} fill className="object-cover" />
                        </div>
                        <span
                          className="text-parea-black font-mono text-base font-medium leading-[150%] tracking-[-0.16px] uppercase"
                          style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace' }}
                        >
                          {profile.name}
                        </span>
                      </div>
                      {profile.count !== null && (
                        <div className="flex px-2 justify-center items-center rounded-full bg-parea-yellow border border-parea-black">
                          <span
                            className="text-parea-black font-mono text-base font-medium leading-[150%] tracking-[-0.16px] uppercase"
                            style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace' }}
                          >
                            {profile.count}
                          </span>
                        </div>
                      )}
                    </div>
                  ))}
              </div>
            </>
          ) : (
            <>
              {/* Notifications */}
              <div className="flex flex-col items-start self-stretch">
                <div className="flex items-center justify-center self-stretch pt-5 pb-3 border-b border-solid border-parea-black">
                  <button
                    onClick={handleMarkAllAsRead}
                    disabled={notifications.length === 0}
                    className="flex items-center justify-center px-5 py-2 text-parea-black font-mono text-base font-medium uppercase tracking-[-0.16px] leading-relaxed cursor-pointer bg-transparent border-0 focus:outline-none whitespace-nowrap"
                    style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace' }}
                  >
                    MARK ALL AS READ
                  </button>
                  <button
                    onClick={() => {
                      void handleDeleteAllNotifications();
                    }}
                    disabled={notifications.length === 0}
                    className="flex items-center justify-center px-5 py-2 text-parea-black font-mono text-base font-medium uppercase tracking-[-0.16px] leading-relaxed cursor-pointer bg-transparent border-0 focus:outline-none whitespace-nowrap"
                    style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace' }}
                  >
                    DELETE ALL
                  </button>
                </div>

                <div className="flex flex-1 px-4 flex-col items-start gap-2 self-stretch overflow-y-auto">
                  {followRequestActionError && (
                    <div className="flex py-2 px-4 items-center self-stretch">
                      <span
                        className="text-parea-black text-sm font-medium leading-[150%]"
                        style={{ fontFamily: 'var(--font-inter), sans-serif' }}
                      >
                        {followRequestActionError}
                      </span>
                    </div>
                  )}
                  {isLoadingNotifications ? (
                    <div className="flex py-4 px-4 items-center self-stretch">
                      <span
                        className="text-parea-black text-sm font-medium leading-[150%]"
                        style={{ fontFamily: 'var(--font-inter), sans-serif' }}
                      >
                        Loading notifications...
                      </span>
                    </div>
                  ) : notificationsError ? (
                    <div className="flex py-4 px-4 items-center self-stretch">
                      <span
                        className="text-parea-black text-sm font-medium leading-[150%]"
                        style={{ fontFamily: 'var(--font-inter), sans-serif' }}
                      >
                        {notificationsError}
                      </span>
                    </div>
                  ) : notifications.length === 0 ? (
                    <div className="flex py-4 px-4 items-center self-stretch">
                      <span
                        className="text-parea-black text-sm font-medium leading-[150%]"
                        style={{ fontFamily: 'var(--font-inter), sans-serif' }}
                      >
                        No notifications yet.
                      </span>
                    </div>
                  ) : (
                    notifications.map((notification) => (
                      <div
                        key={notification.id}
                        className={`flex py-2 px-4 items-center gap-4 self-stretch rounded-lg transition-colors ${notification.read ? 'bg-parea-white' : 'bg-parea-yellow/30'
                          }`}
                      >
                        <div className="flex items-start flex-1">
                          <span
                            className="text-parea-black text-sm font-medium leading-[150%]"
                            style={{ fontFamily: 'var(--font-inter), sans-serif' }}
                          >
                            {formatNotificationText(notification)}
                          </span>
                        </div>
                        {notification.type === 'follow_request' && (
                          <div className="flex items-center gap-2">
                            <button
                              onClick={() => {
                                void handleFollowRequestAction(notification, 'accept');
                              }}
                              disabled={followRequestActionId === notification.id}
                              className="rounded border border-parea-black bg-parea-yellow px-2 py-1 text-xs font-medium uppercase text-parea-black hover:opacity-90 disabled:opacity-60 disabled:cursor-not-allowed"
                            >
                              Accept
                            </button>
                            <button
                              onClick={() => {
                                void handleFollowRequestAction(notification, 'decline');
                              }}
                              disabled={followRequestActionId === notification.id}
                              className="rounded border border-parea-black bg-parea-white px-2 py-1 text-xs font-medium uppercase text-parea-black hover:opacity-90 disabled:opacity-60 disabled:cursor-not-allowed"
                            >
                              Decline
                            </button>
                          </div>
                        )}
                        <button
                          onClick={() => {
                            void handleDeleteNotification(notification.id);
                          }}
                          className="w-6 h-6 flex items-center justify-center hover:opacity-70 transition-opacity cursor-pointer"
                        >
                          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none">
                            <path d="M14.9992 8.99998L8.99922 15M8.99922 8.99998L14.9992 15M3.84922 8.61998C3.70326 7.9625 3.72567 7.27882 3.91437 6.63231C4.10308 5.98581 4.45196 5.39742 4.92868 4.9217C5.40541 4.44597 5.99453 4.09832 6.64142 3.91097C7.28832 3.72362 7.97205 3.70264 8.62922 3.84998C8.99093 3.28428 9.48922 2.81873 10.0782 2.49626C10.6671 2.17379 11.3278 2.00476 11.9992 2.00476C12.6707 2.00476 13.3313 2.17379 13.9203 2.49626C14.5092 2.81873 15.0075 3.28428 15.3692 3.84998C16.0274 3.702 16.7123 3.72288 17.3602 3.91069C18.0081 4.09849 18.598 4.44712 19.0751 4.92413C19.5521 5.40114 19.9007 5.99105 20.0885 6.63898C20.2763 7.28691 20.2972 7.97181 20.1492 8.62998C20.7149 8.99168 21.1805 9.48998 21.5029 10.0789C21.8254 10.6679 21.9944 11.3285 21.9944 12C21.9944 12.6714 21.8254 13.3321 21.5029 13.921C21.1805 14.51 20.7149 15.0083 20.1492 15.37C20.2966 16.0271 20.2756 16.7109 20.0882 17.3578C19.9009 18.0047 19.5532 18.5938 19.0775 19.0705C18.6018 19.5472 18.0134 19.8961 17.3669 20.0848C16.7204 20.2735 16.0367 20.2959 15.3792 20.15C15.018 20.7178 14.5193 21.1854 13.9293 21.5093C13.3394 21.8332 12.6772 22.003 12.0042 22.003C11.3312 22.003 10.669 21.8332 10.0791 21.5093C9.48914 21.1854 8.99045 20.7178 8.62922 20.15C7.97205 20.2973 7.28832 20.2763 6.64142 20.089C5.99453 19.9016 5.40541 19.554 4.92868 19.0783C4.45196 18.6025 4.10308 18.0141 3.91437 17.3676C3.72567 16.7211 3.70326 16.0374 3.84922 15.38C3.27917 15.0192 2.80963 14.5201 2.48426 13.9292C2.1589 13.3382 1.98828 12.6746 1.98828 12C1.98828 11.3254 2.1589 10.6617 2.48426 10.0708C2.80963 9.4798 3.27917 8.98073 3.84922 8.61998Z" stroke="black" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                          </svg>
                        </button>
                      </div>
                    ))
                  )}
                </div>
              </div>
            </>
          )}

          {/* Back Button */}
          {activeDrawer && (
            <div className="absolute -left-18 top-0 w-18 h-18 flex items-center justify-center bg-parea-black">
              <SidebarToggle isOpen={true} onClick={handleDrawerClose} />
            </div>
          )}
        </aside>
    </div>
  );
}
