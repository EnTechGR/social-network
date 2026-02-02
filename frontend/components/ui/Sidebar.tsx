// components/ui/Sidebar.tsx
'use client';

import React, { useState, useRef, useEffect } from 'react';
import { MessageCircle, Bell, LogOut, CircleX } from 'lucide-react';
import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import SidebarToggle from './SidebarToggle';
import Tabs from './Tabs';
import Button from './Button';
import { logout } from '@/lib/api';
import { clearAuth } from '@/lib/auth';

type DrawerType = 'chat' | 'notifications' | null;

export default function Sidebar() {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [activeDrawer, setActiveDrawer] = useState<DrawerType>(null);

  const handleToggle = () => {
    setIsOpen(!isOpen);
    if (!isOpen) {
      setActiveDrawer(null); // Close drawer when closing sidebar
    }
  };

  const handleDrawerOpen = (type: 'chat' | 'notifications') => {
    setIsOpen(false); // Close main sidebar, return to default (icons only)
    setActiveDrawer(type); // Open step-2 drawer
  };

  const handleDrawerClose = () => {
    setActiveDrawer(null);
  };

  const handleLogout = async () => {
    try {
      // Call logout API to clear session on backend
      await logout();
    } catch (error) {
      console.error('Logout API call failed:', error);
      // Continue with client-side logout even if API fails
    } finally {
      // Clear client-side auth data
      clearAuth();
      // Redirect to login page
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

  // ============ DEFAULT SIDEBAR (ICONS ONLY) - Shows when drawer is open OR when both closed ============
  if (!isOpen) {
    return (
      <div ref={sidebarRef}>
        <aside
          className="
            fixed
            left-0
            top-0
            h-screen
            w-[72px]
            z-50
            flex
            items-start
            bg-parea-black
            border-r
            border-parea-black
          "
        >
          <div
            className="
              flex
              flex-col
              items-center
              gap-8
              flex-1
              self-stretch
              py-6
              px-3
            "
          >
            <SidebarToggle isOpen={isOpen} onClick={handleToggle} />

            <div
              className="
                flex
                flex-col
                items-center
                gap-1
                flex-1
                self-stretch
              "
            >
              <button
                onClick={() => handleDrawerOpen('chat')}
                className="p-2 rounded-lg transition-colors group cursor-pointer"
                aria-label="Messages"
              >
                <MessageCircle 
                  className={`w-6 h-6 transition-colors ${
                    activeDrawer === 'chat' 
                      ? 'text-parea-yellow' 
                      : 'text-parea-white group-hover:text-parea-yellow'
                  }`} 
                />
              </button>

              <button
                onClick={() => handleDrawerOpen('notifications')}
                className="p-2 rounded-lg transition-colors group cursor-pointer"
                aria-label="Notifications"
              >
                <Bell 
                  className={`w-6 h-6 transition-colors ${
                    activeDrawer === 'notifications' 
                      ? 'text-parea-yellow' 
                      : 'text-parea-white group-hover:text-parea-yellow'
                  }`} 
                />
              </button>
            </div>

            <div
              className="
                w-10
                h-10
                rounded-full
                border-2
                border-parea-yellow
                overflow-hidden
                relative
              "
            >
              <Image
                src="/test-avatar.png"
                alt="User avatar"
                fill
                className="object-cover"
              />
            </div>

            <button
              onClick={handleLogout}
              className="p-2 rounded-lg transition-colors group cursor-pointer"
              aria-label="Logout"
            >
              <LogOut className="w-6 h-6 text-parea-white group-hover:text-parea-yellow transition-colors" />
            </button>
          </div>
        </aside>

        {/* Step-2 Drawer (Chat or Notifications) - Only shows when activeDrawer is set */}
        {activeDrawer && (
          <aside
            className="
              fixed
              left-[72px]
              top-0
              h-screen
              w-[315px]
              z-40
              flex
              flex-col
              items-start
              self-stretch
              bg-parea-white
              border-r
              border-parea-black
            "
          >
            {activeDrawer === 'chat' ? (
              <>
                {/* Tabs Section */}
                <div
                  className="
                    flex
                    pt-5
                    px-2
                    pb-3
                    justify-center
                    items-center
                    self-stretch
                  "
                >
                  <Tabs
                    tabs={['DIRECT CHATS', 'GROUP CHATS']}
                    defaultTab="DIRECT CHATS"
                    className="w-full justify-center"
                  />
                </div>

                {/* Chat List */}
                <div
                  className="
                    flex
                    h-[820px]
                    pt-2
                    flex-col
                    items-start
                    gap-4
                    self-stretch
                    overflow-y-auto
                  "
                >
                  <div
                    className="
                      flex
                      h-[820px]
                      pt-2
                      flex-col
                      items-start
                      gap-4
                      self-stretch
                      px-4
                    "
                  >
                    {/* Sample Chat Profile Items - Replace with dynamic data */}
                    {[
                      { name: 'JENNIFER WHITE', count: 2, avatar: '/test-avatar.png' },
                      { name: 'ALEX DONHAM', count: 1, avatar: '/test-avatar.png' },
                      { name: 'KAREN HILLS', count: null, avatar: '/test-avatar.png' },
                    ].map((profile, index) => (
                      <div
                        key={index}
                        className="
                          flex
                          h-14
                          py-2
                          px-4
                          justify-between
                          items-center
                          self-stretch
                          rounded-lg
                          hover:bg-parea-grey/30
                          transition-colors
                          cursor-pointer
                        "
                      >
                        {/* Avatar + Name */}
                        <div className="flex items-center gap-3">
                          <div
                            className="
                              w-10
                              h-10
                              rounded-full
                              overflow-hidden
                              relative
                              shrink-0
                            "
                          >
                            <Image
                              src={profile.avatar}
                              alt={profile.name}
                              fill
                              className="object-cover"
                            />
                          </div>
                          <span
                            className="
                              text-parea-black
                              font-mono
                              text-base
                              font-medium
                              leading-[150%]
                              tracking-[-0.16px]
                              uppercase
                            "
                            style={{
                              fontFamily: 'var(--font-ibm-plex-mono), monospace',
                            }}
                          >
                            {profile.name}
                          </span>
                        </div>

                        {/* Count Badge */}
                        {profile.count !== null && (
                          <div
                            className="
                              flex
                              px-2
                              justify-center
                              items-center
                              rounded-full
                              bg-parea-yellow
                              border
                              border-parea-black
                            "
                          >
                            <span
                              className="
                                text-parea-black
                                font-mono
                                text-base
                                font-medium
                                leading-[150%]
                                tracking-[-0.16px]
                                uppercase
                              "
                              style={{
                                fontFamily: 'var(--font-ibm-plex-mono), monospace',
                              }}
                            >
                              {profile.count}
                            </span>
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              </>
            ) : (
              <>
                {/* Notifications Outer Frame */}
                <div
                  className="
                    flex
                    w-[315px]
                    flex-col
                    items-start
                    self-stretch
                  "
                >
                  {/* First div - Tertiary Buttons */}
                  <div
                    className="
                      flex
                      w-[315px]
                      flex-row
                      items-start
                      self-stretch
                      pt-5
                      px-4
                      pb-3
                      justify-between
                    "
                  >
                    <Button variant="tertiary" onClick={() => {}}>
                      MARK ALL READ
                    </Button>
                    <Button variant="tertiary" onClick={() => {}}>
                      DELETE ALL
                    </Button>
                  </div>

                  {/* Notifications List */}
                  <div
                    className="
                      flex
                      h-[813px]
                      px-4
                      flex-col
                      items-start
                      gap-2
                      self-stretch
                      overflow-y-auto
                    "
                  >
                    {/* Sample Notification Items - Replace with dynamic data */}
                    {[
                      { text: "Jennifer White liked your post 'Vegan cheesecake in 20 minutes'", unread: true },
                      { text: "Alex Donham commented on your post", unread: true },
                      { text: "Karen Hills shared your post", unread: false },
                      { text: "Anette Black liked your photo", unread: false },
                    ].map((notification, index) => (
                      <div
                        key={index}
                        className={`
                          flex
                          py-2
                          px-4
                          items-center
                          gap-4
                          self-stretch
                          rounded-lg
                          transition-colors
                          ${notification.unread ? 'bg-parea-yellow/20' : 'bg-parea-white'}
                        `}
                      >
                        {/* Text div */}
                        <div
                          className="
                            flex
                            items-start
                            flex-1
                          "
                        >
                          <span
                            className="
                              text-parea-black
                              text-sm
                              font-medium
                              leading-[150%]
                            "
                            style={{
                              fontFamily: 'var(--font-inter), sans-serif',
                              fontSize: '14px',
                              fontWeight: 500,
                              lineHeight: '150%',
                            }}
                          >
                            {notification.text}
                          </span>
                        </div>

                        {/* CircleX Icon */}
                        <button
                          className="
                            w-6
                            h-6
                            flex
                            items-center
                            justify-center
                            hover:opacity-70
                            transition-opacity
                            cursor-pointer
                          "
                        >
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
                absolute
                left-[-72px]
                top-0
                w-[72px]
                h-[72px]
                flex
                items-center
                justify-center
                bg-parea-black
                border-r
                border-parea-black
                hover:bg-parea-black/90
                transition-colors
                cursor-pointer
              "
            >
              <div className="w-9 h-9 rounded-full border border-parea-yellow flex items-center justify-center">
                <svg
                  width="15"
                  height="15"
                  viewBox="0 0 15 15"
                  fill="none"
                  className="rotate-180"
                >
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

  // ============ OPEN STATE (DRAWER WITH TEXT) ============
  return (
    <div ref={sidebarRef}>
    <aside
      className="
        fixed
        left-0
        top-0
        h-screen
        w-[315px]
        z-50
        flex
        flex-col
        items-start
        border-r
        border-parea-black
        bg-parea-black
      "
    >
      {/* Header - Logo + Toggle */}
      <div
        className="
          flex
          h-[91px]
          pt-6
          px-4
          pb-4
          flex-col
          items-start
          gap-2.5
          shrink-0
          self-stretch
        "
      >
        <div className="flex w-full justify-between items-center">
          <Image
            src="/logo-light.svg"
            alt="Parea Logo"
            width={80}
            height={32}
            className="h-8 w-auto"
          />
          <SidebarToggle isOpen={isOpen} onClick={handleToggle} />
        </div>
      </div>

      {/* Menu Container */}
      <div
        className="
          flex
          flex-col
          justify-between
          items-start
          flex-1
          self-stretch
        "
      >
        {/* Menu List (top) */}
        <div
          className="
            flex
            pt-2
            px-4
            pb-0
            flex-col
            items-start
            self-stretch
          "
        >
          {/* Chat item */}
          <button
            onClick={() => handleDrawerOpen('chat')}
            className="
              flex
              p-2
              items-center
              gap-2
              self-stretch
              rounded-lg
              transition-colors
              group
              cursor-pointer
            "
          >
            <MessageCircle className="w-6 h-6 text-parea-white group-hover:text-parea-yellow transition-colors" />
            <span className="text-parea-white font-mono text-sm font-medium uppercase tracking-wide group-hover:text-parea-yellow transition-colors">
              Chats
            </span>
          </button>

          {/* Notifications item */}
          <button
            onClick={() => handleDrawerOpen('notifications')}
            className="
              flex
              p-2
              items-center
              gap-2
              self-stretch
              rounded-lg
              transition-colors
              group
              justify-between
              cursor-pointer
            "
          >
            <div className="flex items-center gap-2">
              <Bell className="w-6 h-6 text-parea-white group-hover:text-parea-yellow transition-colors" />
              <span className="text-parea-white font-mono text-sm font-medium uppercase tracking-wide group-hover:text-parea-yellow transition-colors">
                Notifications
              </span>
            </div>
            <span
              className="
                flex
                items-center
                justify-center
                min-w-[24px]
                h-6
                px-2
                rounded-full
                border
                border-parea-yellow
                text-parea-yellow
                font-mono
                text-xs
                font-medium
              "
            >
              24
            </span>
          </button>
        </div>

        {/* Menu Bottom (profile + logout) */}
        <div
          className="
            flex
            px-4
            pb-8
            flex-col
            items-start
            gap-2
            self-stretch
          "
        >
          {/* Profile item */}
          <Link
            href="/profile"
            className="
              flex
              px-2
              items-center
              gap-3
              self-stretch
              rounded-lg
              transition-colors
              group
              cursor-pointer
            "
          >
            <div
              className="
                w-10
                h-10
                rounded-full
                border-2
                border-parea-yellow
                overflow-hidden
                relative
                shrink-0
            "
            >
              <Image
                src="/test-avatar.png"
                alt="User avatar"
                fill
                className="object-cover"
              />
            </div>
            <span className="text-parea-white font-mono text-sm font-medium uppercase tracking-wide group-hover:text-parea-yellow transition-colors">
              Profile
            </span>
          </Link>

          {/* Logout item */}
          <button
            onClick={handleLogout}
            className="
              flex
              px-2
              items-center
              gap-3
              self-stretch
              rounded-lg
              transition-colors
              group
              cursor-pointer
            "
          >
            <div className="p-2">
              <LogOut className="w-6 h-6 text-parea-white group-hover:text-parea-yellow transition-colors" />
            </div>
            <span className="text-parea-white font-mono text-sm font-medium uppercase tracking-wide group-hover:text-parea-yellow transition-colors">
              Logout
            </span>
          </button>
        </div>
      </div>
    </aside>
    </div>
  );
}