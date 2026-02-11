/**
 * components/ui/Tabs.tsx
 *
 * Parea Tabs Component
 * Tab options: Posts, Events, Users, Groups (or custom)
 */

'use client';

import { useState, useRef, useEffect } from 'react';

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
  const [sliderStyle, setSliderStyle] = useState({});
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const index = tabs.indexOf(activeTab);
    const tab = tabRefs.current[index];
    const container = containerRef.current;

    if (tab && container) {
      const tabRect = tab.getBoundingClientRect();
      const containerRect = container.getBoundingClientRect();
      setSliderStyle({
        '--slider-left': `${tabRect.left - containerRect.left}px`,
        '--slider-width': `${tabRect.width}px`,
      } as React.CSSProperties);
    }
  }, [activeTab, tabs]);

  const handleTabClick = (tab: string) => {
    setActiveTab(tab);
    if (onTabChange) {
      onTabChange(tab);
    }
  };

  // Frame styles
  const frameStyles = `
    inline-flex
    flex-wrap
    p-1
    items-start
    gap-1
    rounded-full
    border
    border-parea-border
    md:border
    border-transparent
    md:border-parea-border
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
    cursor-pointer
  `;

  return (
    <div ref={containerRef} className={`${frameStyles} ${className} relative`} style={sliderStyle}>
      {/* Sliding background */}
      <div
        className="absolute rounded-full border border-parea-black bg-parea-yellow transition-all duration-250 ease-out"
        style={{
          left: 'var(--slider-left, 0)',
          width: 'var(--slider-width, 0)',
          top: '4px',
          bottom: '4px',
        }}
      />

      {tabs.map((tab, i) => (
        <button
          key={tab}
          ref={(el) => { tabRefs.current[i] = el; }}
          type="button"
          onClick={() => handleTabClick(tab)}
          className={`${baseTabStyles} relative z-10 border border-transparent bg-transparent`}
        >
          {tab}
        </button>
      ))}
    </div>
  );
}