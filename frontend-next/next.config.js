/** @type {import('next').NextConfig} */
const backendUrl = (process.env.GODNSLOG_API_URL || 'http://localhost:8080').replace(/\/$/, '')

const nextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  distDir: 'dist',
  images: {
    unoptimized: true,
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${backendUrl}/api/:path*`,
      },
    ]
  },
}

module.exports = nextConfig
