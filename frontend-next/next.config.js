/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  async rewrites() {
    const apiUrl = process.env.GODNSLOG_API_URL || 'http://localhost:8080'
    return [
      {
        source: '/api/v2/:path*',
        destination: `${apiUrl}/api/v2/:path*`,
      },
    ]
  },
}

module.exports = nextConfig
