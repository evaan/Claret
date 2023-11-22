import type { Metadata } from 'next'
import Providers from './providers'

export const metadata: Metadata = {
  title: 'Claret Schedule Builder',
  description: 'Schedule Builder for MUN Students',
}

export default function RootLayout({children,}: {children: React.ReactNode}) {
  return(
    <html lang="en">
      <body>
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  )
}