'use client';

import Link from 'next/link';
import SlidingEaseVerticalBars from '@/components/ui/SlidingEaseVerticalBars';

export default function CatchAllNotFound() {
  return (
    <div className="relative w-full h-[calc(100vh-72px)]">
      <SlidingEaseVerticalBars backgroundColor="#FFFFFF" />

      <div className="relative z-10 flex flex-col items-center justify-center gap-8 h-full pointer-events-none">
        <p
          className="font-bold leading-tight text-[128px] text-parea-black"
          style={{ fontFamily: 'var(--font-inter), sans-serif' }}
        >
          404
        </p>

        <p
          className="font-bold leading-tight text-h2 text-parea-black"
          style={{ fontFamily: 'var(--font-inter), sans-serif' }}
        >
          Page Not Found
        </p>

        <Link
          href="/feed"
          className="pointer-events-auto flex items-center justify-center gap-3 px-5 py-2 bg-parea-black border border-parea-black text-white uppercase tracking-[-0.16px] text-regular leading-relaxed"
          style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace', fontWeight: 500 }}
        >
          BACK TO FEED
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M5 12H19M19 12L13 6M19 12L13 18" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </Link>
      </div>
    </div>
  );
}
