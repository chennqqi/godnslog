/** @type {import('next').NextConfig} */
const backendUrl = (process.env.GODNSLOG_API_URL || 'http://localhost:8080').replace(/\/$/, '')

const nextConfig = {
  reactStrictMode: true,
  distDir: 'dist',
  images: {
    unoptimized: true,
  },
  async rewrites() {
    return [
      {
        source: '/api/v2/:path*',
        destination: `${backendUrl}/api/v2/:path*`,
      },
      {
        source: '/api/v1/:path*',
        destination: `${backendUrl}/api/v1/:path*`,
      },
    ]
  },
}

module.exports = nextConfig
