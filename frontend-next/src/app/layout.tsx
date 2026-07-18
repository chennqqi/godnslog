import type { Metadata } from 'next'
import './globals.css'
import { QueryProvider } from '@/components/query-provider'
import { I18nProvider } from '@/lib/i18n-context'

export const dynamic = 'force-dynamic'

export const metadata: Metadata = {
  title: 'GODNSLOG 2.0 - OAST Interaction Verification Platform',
  description: 'OAST interaction verification and evidence platform',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: `(function(){try{var t=localStorage.getItem('godnslog-theme');var m=t?JSON.parse(t).state.theme:'system';var d=m==='dark'||(m==='system'&&window.matchMedia('(prefers-color-scheme:dark)').matches);if(d)document.documentElement.classList.add('dark');}catch(e){}})();`,
          }}
        />
      </head>
      <body className="font-sans antialiased">
        <I18nProvider>
          <QueryProvider>{children}</QueryProvider>
        </I18nProvider>
      </body>
    </html>
  )
}
