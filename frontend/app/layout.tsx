import type { Metadata } from "next";
import { IBM_Plex_Mono, Inter } from "next/font/google";
import "./globals.css";

// Inter - for body text and headings
const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700", "800"],
});

// IBM Plex Mono - for labels, buttons, special text
const ibmPlexMono = IBM_Plex_Mono({
  variable: "--font-ibm-plex-mono",
  subsets: ["latin"],
  weight: ["400", "500", "600", "700"],
});

export const metadata: Metadata = {
  title: "Parea Social Network",
  description: "Connect with your community",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  // Noise texture - mono, high density for paper-like texture, color #000 at 10% opacity
  const noiseStyle = {
    backgroundImage: `url("data:image/svg+xml,%3Csvg width='200' height='200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='5' stitchTiles='stitch'/%3E%3CfeColorMatrix type='saturate' values='0'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)'/%3E%3C/svg%3E")`,
    opacity: 0.18,
    mixBlendMode: 'normal' as const,
  };

  return (
    <html lang="en">
      <body className={`${inter.variable} ${ibmPlexMono.variable} antialiased relative`}>
        {/* Global noise overlay - applies to all pages */}
        <div
          className="fixed inset-0 pointer-events-none z-9999"
          style={noiseStyle}
        />
        {children}
      </body>
    </html>
  );
}
