"use client";
import { usePathname } from 'next/navigation';
import Navbar from './navbar';

export default function NavbarWrapper() {
  const pathname = usePathname();
  const hideNavbar = pathname === '/login' || pathname === '/signup';
  if (hideNavbar) return null;
  return <Navbar />;
}
