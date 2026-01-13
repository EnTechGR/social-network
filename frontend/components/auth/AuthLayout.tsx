/**
 * components/auth/AuthLayout.tsx
 *
 * Shared layout for authentication pages (login, signup).
 * Features a split-screen design with gradient left panel and form right panel.
 */

import Image from 'next/image';

interface AuthLayoutProps {
  children: React.ReactNode;
}

export default function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className="flex flex-col md:flex-row h-screen">
      {/* Left Panel - Gradient with dots pattern */}
      <div className="relative w-full md:w-[40%] h-64 md:h-full bg-linear-to-b from-parea-yellow to-parea-black overflow-hidden">
        {/* Dot Pattern Overlay */}
        <div 
          className="absolute inset-0 opacity-100 z-0"
          style={{
            backgroundImage: 'radial-gradient(circle, #121214 1px, transparent 1px)',
            backgroundSize: '20px 20px',
            backgroundPosition: '0 0',
          }}
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
