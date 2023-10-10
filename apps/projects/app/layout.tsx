import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import { Box } from 'ui';
import { Sidebar } from './components/shared';
import './styles/global.css';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'Complexus Projects',
  description: 'Complexus Projects',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}): JSX.Element {
  return (
    <html lang='en' suppressHydrationWarning>
      <body className={inter.className}>
        <main className='grid grid-cols-[220px_auto] h-screen'>
          <Sidebar />
          <Box>{children}</Box>
        </main>
      </body>
    </html>
  );
}
