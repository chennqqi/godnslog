import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

/**
 * Authentication guard middleware.
 * Redirects unauthenticated requests to /login.
 * Allows public routes (login, api auth) to pass through.
 */
export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Public routes that do not require authentication
  const publicPaths = ['/login', '/api/auth']
  const isPublicPath = publicPaths.some((p) => pathname.startsWith(p))

  if (isPublicPath) {
    return NextResponse.next()
  }

  // Check for auth token in cookies or headers
  const token =
    request.cookies.get('token')?.value ||
    request.headers.get('authorization')?.replace('Bearer ', '')

  if (!token || token.length === 0) {
    // Redirect to login page for unauthenticated requests
    const loginUrl = new URL('/login', request.url)
    return NextResponse.redirect(loginUrl)
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|.*\\.png$|.*\\.svg$).*)',
  ],
}
