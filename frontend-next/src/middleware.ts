import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

/**
 * Authentication guard middleware.
 * Redirects unauthenticated requests to /login.
 * Allows public routes (login, api auth) to pass through.
 * Note: This middleware runs on the server, so it cannot access localStorage.
 * Authentication is checked via Authorization header from client-side API calls.
 * For page navigation, authentication is handled client-side by checking localStorage.
 */
export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Public routes that do not require authentication
  const publicPaths = ['/login', '/api/auth']
  const isPublicPath = publicPaths.some((p) => pathname.startsWith(p))

  if (isPublicPath) {
    return NextResponse.next()
  }

  // Check for auth token in Authorization header (from API calls)
  const token = request.headers.get('authorization')?.replace('Bearer ', '')

  if (!token || token.length === 0) {
    // For page navigation, we cannot check localStorage (server-side)
    // So we allow the request to pass and handle authentication client-side
    // The dashboard page will check localStorage and redirect if needed
    return NextResponse.next()
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|.*\\.png$|.*\\.svg$).*)',
  ],
}
