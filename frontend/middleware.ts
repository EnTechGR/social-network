import { NextRequest, NextResponse } from 'next/server';

const publicRoutes = ['/login', '/signup'];

const skipAuth = false;

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Allow public routes
  if (publicRoutes.some((route) => pathname.startsWith(route))) {
    return NextResponse.next();
  }

  // Skip auth check during development
  if (skipAuth) {
    return NextResponse.next();
  }

  // Check for session cookie
  const sessionCookie = request.cookies.get('id');

  if (!sessionCookie) {
    const loginUrl = new URL('/login', request.url);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all routes except:
     * - _next/static (static files)
     * - _next/image (image optimization)
     * - favicon.ico, icon.png, etc.
     * - public assets
     */
    '/((?!_next/static|_next/image|favicon.ico|icon.png|logo.svg|logo-light.svg|.*\\.png$|.*\\.jpg$).*)',
  ],
};
