/**
 * components/ui/Tabs.tsx
 *
 * Parea Tabs Component
 * Tab options: Posts, Events, Users, Groups (or custom)
 */

'use client';

import { useState } from 'react';

interface TabsProps {
  tabs: string[];
  defaultTab?: string;
  onTabChange?: (tab: string) => void;
  className?: string;
}

export default function Tabs({
  tabs,
  defaultTab,
  onTabChange,
  className = '',
}: TabsProps) {
  const [activeTab, setActiveTab] = useState(defaultTab || tabs[0]);

  const handleTabClick = (tab: string) => {
    setActiveTab(tab);
    if (onTabChange) {
      onTabChange(tab);
    }
  };

  // Frame styles
  const frameStyles = `
    inline-flex
    p-1
    items-start
    gap-1
    rounded-full
    border
    border-parea-border
  `;

  // Base tab option styles (shared)
  const baseTabStyles = `
    flex
    flex-col
    items-start
    py-2
    px-5
    gap-2
    rounded-full
    font-display
    text-[15px]
    font-medium
    leading-[1.5]
    tracking-tight
    uppercase
    text-parea-black
    transition-all
    duration-200
    cursor-pointer
  `;

  // Active tab styles
  const activeTabStyles = `
    border
    border-parea-black
    bg-parea-yellow
  `;

  // Default tab styles
  const defaultTabStyles = `
    border
    border-transparent
    bg-transparent
    hover:bg-parea-grey/30
  `;

  return (
    <div className={`${frameStyles} ${className}`}>
      {tabs.map((tab) => (
        <button
          key={tab}
          type="button"
          onClick={() => handleTabClick(tab)}
          className={`
            ${baseTabStyles}
            ${activeTab === tab ? activeTabStyles : defaultTabStyles}
          `}
        >
          {tab}
        </button>
      ))}
    </div>
  );
}