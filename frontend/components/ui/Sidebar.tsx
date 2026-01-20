// components/ui/Sidebar.tsx
'use client';

import React, { useState } from 'react';
import { MessageCircle, Bell, LogOut } from 'lucide-react';
import Image from 'next/image';
import SidebarToggle from './SidebarToggle';

export default function Sidebar() {
  const [isOpen, setIsOpen] = useState(false);

  const handleToggle = () => {
    setIsOpen(!isOpen);
  };

  return (
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
      {/* Container level 1 - column layout */}
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
        {/* Toggle button at top */}
        <SidebarToggle isOpen={isOpen} onClick={handleToggle} />

        {/* Container level 2 - icons section (grows to push avatar/logout down) */}
        <div
          className="
            flex
            flex-col
            items-center
            gap-8
            flex-1
            self-stretch
          "
        >
          {/* Comment icon */}
          <button
            className="
              p-2
              rounded-lg
              transition-colors
              group
            "
            aria-label="Messages"
          >
            <MessageCircle className="w-6 h-6 text-parea-white group-hover:text-parea-yellow transition-colors" />
          </button>

          {/* Notification bell icon */}
          <button
            className="
              p-2
              rounded-lg
              transition-colors
              group
            "
            aria-label="Notifications"
          >
            <Bell className="w-6 h-6 text-parea-white group-hover:text-parea-yellow transition-colors" />
          </button>
        </div>

        {/* Avatar with parea yellow border */}
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
            src="/test-avatar.jpg"
            alt="User avatar"
            fill
            className="object-cover"
          />
        </div>

        {/* Logout icon */}
        <button
          className="
            p-2
            rounded-lg
            transition-colors
            group
          "
          aria-label="Logout"
        >
          <LogOut className="w-6 h-6 text-parea-white group-hover:text-parea-yellow transition-colors" />
        </button>
      </div>
    </aside>
  );
}