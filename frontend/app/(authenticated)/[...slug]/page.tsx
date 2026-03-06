'use client';

import { useEffect, useState } from 'react';
import Script from 'next/script';
import Link from 'next/link';
import { DOT_COLORS } from '@/app/not-found-config';

export default function CatchAllNotFound() {
  const [scriptsReady, setScriptsReady] = useState(false);

  useEffect(() => {
    if (!scriptsReady || typeof window === 'undefined' || !(window as unknown as { gsap?: { registerPlugin: (p: unknown) => void; utils: { random: (a: number, b: number, c?: number) => number }; set: (el: HTMLElement, props: object) => void } }).gsap) return;

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const gsap = (window as unknown as { gsap: any }).gsap;
    const Physics2DPlugin = (window as unknown as { Physics2DPlugin?: unknown }).Physics2DPlugin;
    if (!Physics2DPlugin) return;

    gsap.registerPlugin(Physics2DPlugin);

    const handleClick = (event: MouseEvent) => {
      const dotCount = Math.floor(gsap.utils.random(3, 10, 1) as number);
      for (let i = 0; i < dotCount; i++) {
        const dot = document.createElement('div');
        dot.classList.add('not-found-dot');
        document.body.appendChild(dot);

        gsap.set(dot, {
          backgroundColor: gsap.utils.random(DOT_COLORS) as string,
          top: event.clientY,
          left: event.clientX,
          scale: 0,
        });

        gsap
          .timeline({
            onComplete: () => dot.remove(),
          })
          .to(dot, {
            scale: gsap.utils.random(0.5, 1) as number,
            duration: 0.3,
            ease: 'power3.out',
          })
          .to(
            dot,
            {
              duration: 2,
              physics2D: {
                velocity: gsap.utils.random(200, 650) as number,
                angle: gsap.utils.random(0, 360) as number,
                gravity: 500,
              },
              autoAlpha: 0,
              ease: 'none',
            },
            '<'
          );
      }
    };

    document.body.style.overflow = 'clip';
    document.addEventListener('click', handleClick);
    return () => {
      document.removeEventListener('click', handleClick);
      document.body.style.overflow = '';
    };
  }, [scriptsReady]);

  return (
    <>
      <Script
        src="https://cdn.jsdelivr.net/npm/gsap@3.13.0/dist/gsap.min.js"
        strategy="afterInteractive"
        onLoad={() => {
          (window as unknown as { _gsapLoaded?: boolean })._gsapLoaded = true;
        }}
      />
      <Script
        src="https://cdn.jsdelivr.net/npm/gsap@3.13.0/dist/Physics2DPlugin.min.js"
        strategy="afterInteractive"
        onLoad={() => setScriptsReady(true)}
      />

      <div className="relative w-full h-[calc(100vh-72px)] flex bg-white overflow-hidden">
        <div className="relative z-10 flex flex-1 flex-col items-center justify-center gap-8 pointer-events-none">
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
            className="pointer-events-auto rounded-button border border-parea-black bg-[#DDFF30] hover:bg-[#C8E82A] flex items-center justify-center gap-2 px-6 py-3 h-input text-regular text-parea-black uppercase tracking-[-0.16px] transition-colors"
            style={{ fontFamily: 'var(--font-ibm-plex-mono), monospace', fontWeight: 500 }}
          >
            BACK TO FEED
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M5 12H19M19 12L13 6M19 12L13 18" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
          </Link>
        </div>
      </div>
    </>
  );
}
