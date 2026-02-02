/**
 * components/auth/AuthLayout.tsx
 *
 * Shared layout for authentication pages (login, signup).
 * Features a split-screen design with gradient left panel and form right panel.
 */

import Image from 'next/image';
import InteractiveDots from '@/components/ui/InteractiveDots';

interface AuthLayoutProps {
  children: React.ReactNode;
}

export default function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className="flex flex-col md:flex-row h-screen">
      {/* Left Panel - parea yellow to black gradient, interactive dots, logo */}
      <div className="relative w-full md:w-[40%] h-64 md:h-full bg-linear-to-b from-parea-yellow to-parea-black overflow-hidden">
        {/* Interactive dots (transparent bg so gradient shows through) */}
        <InteractiveDots
          backgroundColor="transparent"
          dotColor="#121214"
          gridSpacing={20}
          removeWaveLine
          animationSpeed={0.005}
        />

        {/* Logo */}
        <div className="absolute bottom-8 left-8 z-10">
          <Image
            src="/logo.svg"
            alt="Parea"
            width={173}
            height={50}
            priority
          />
        </div>
      </div>

      {/* Right Panel - Form Area */}
      <div className="flex-1 bg-parea-white relative overflow-y-auto">
        <div className="relative z-10 h-full flex items-center justify-center">
          {children}
        </div>
      </div>
    </div>
  );
}
