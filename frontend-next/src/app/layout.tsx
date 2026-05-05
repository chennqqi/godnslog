import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'GODNSLOG 2.0 - OAST Interaction Verification Platform',
  description: 'OAST交互验证与证据平台',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="zh-CN">
      <body className="font-sans antialiased">{children}</body>
    </html>
  )
}
