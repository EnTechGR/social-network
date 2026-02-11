// components/ui/Sidebar.tsx
'use client';

import { useState, useRef, useEffect } from 'react';
import { MessageCircle, Bell, LogOut, CircleX } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import SidebarToggle from './SidebarToggle';
import Tabs from './Tabs';
import Button from './Button';
import { logout, getProfile, getAvatarUrl } from '@/lib/api';
import { clearAuth } from '@/lib/auth';

type DrawerType = 'chat' | 'notifications' | null;

const DEFAULT_AVATAR = '/user-avatar-default.png';

export default function Sidebar() {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [activeDrawer, setActiveDrawer] = useState<DrawerType>(null);
  const [avatarUrl, setAvatarUrl] = useState(DEFAULT_AVATAR);

  useEffect(() => {
    getProfile()
      .then((p) => {
        const path = p?.avatar?.file_path || p?.avatar?.thumbnail_path;
        setAvatarUrl(getAvatarUrl(path));
      })
      .catch(() => { });
  }, []);

  const handleToggle = () => {
    setIsOpen(!isOpen);
    if (isOpen) {
      setActiveDrawer(null);
    }
  };

  const handleDrawerOpen = (type: 'chat' | 'notifications') => {
    setIsOpen(false);
    setActiveDrawer(type);
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

  const sidebarRef = useRef<HTMLDivElement>(null);

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
            <SidebarToggle isOpen={isOpen} onClick={handleToggle} />
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
                24
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
      {activeDrawer && (
        <aside
          className="
            fixed left-18 top-0 h-screen w-79 z-40
            flex flex-col items-start
            bg-parea-white border-r border-parea-black
          "
        >
          {activeDrawer === 'chat' ? (
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
              <div className="flex h-205 pt-2 flex-col items-start gap-4 self-stretch overflow-y-auto px-4">
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
                <div className="flex items-start self-stretch pt-5 px-4 pb-3 justify-between">
                  <Button variant="tertiary" onClick={() => { }}>
                    MARK ALL READ
                  </Button>
                  <Button variant="tertiary" onClick={() => { }}>
                    DELETE ALL
                  </Button>
                </div>

                <div className="flex h-203 px-4 flex-col items-start gap-2 self-stretch overflow-y-auto">
                  {[
                    { text: "Jennifer White liked your post 'Vegan cheesecake in 20 minutes'", unread: true },
                    { text: 'Alex Donham commented on your post', unread: true },
                    { text: 'Karen Hills shared your post', unread: false },
                    { text: 'Anette Black liked your photo', unread: false },
                  ].map((notification, index) => (
                    <div
                      key={index}
                      className={`flex py-2 px-4 items-center gap-4 self-stretch rounded-lg transition-colors ${notification.unread ? 'bg-parea-yellow/20' : 'bg-parea-white'
                        }`}
                    >
                      <div className="flex items-start flex-1">
                        <span
                          className="text-parea-black text-sm font-medium leading-[150%]"
                          style={{ fontFamily: 'var(--font-inter), sans-serif' }}
                        >
                          {notification.text}
                        </span>
                      </div>
                      <button className="w-6 h-6 flex items-center justify-center hover:opacity-70 transition-opacity cursor-pointer">
                        <CircleX className="w-6 h-6 text-parea-black" />
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </>
          )}

          {/* Back Button */}
          <button
            onClick={handleDrawerClose}
            className="
              absolute -left-18 top-0 w-18 h-18
              flex items-center justify-center
              bg-parea-black border-r border-parea-black
              hover:bg-parea-black/90 transition-colors cursor-pointer
            "
          >
            <div className="w-9 h-9 rounded-full border border-parea-yellow flex items-center justify-center">
              <svg width="16" height="16" viewBox="0 0 15 15" fill="none" className="rotate-180">
                <path
                  d="M0.5 7.5H14.5M14.5 7.5L7.5 0.5M14.5 7.5L7.5 14.5"
                  stroke="var(--parea-yellow)"
                  strokeWidth="1"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </div>
          </button>
        </aside>
      )}
    </div>
  );
}
