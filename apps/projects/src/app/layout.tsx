import type { Metadata } from "next";
import { Instrument_Sans as InstrumentSans } from "next/font/google";
import { type ReactNode } from "react";
import "../styles/global.css";
import { SessionProvider } from "next-auth/react";
import { auth } from "@/auth";
import { Providers } from "./providers";
import { Toaster } from "./toaster";
import { OnlineStatusMonitor } from "./online-monitor";

const font = InstrumentSans({
  subsets: ["latin"],
  display: "swap",
  weight: ["400", "500", "600", "700"],
});

export const metadata: Metadata = {
  title: "Complexus",
  viewport: {
    width: "device-width",
    initialScale: 1,
    maximumScale: 1,
    userScalable: false,
  },
};

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await auth();

  return (
    <html className={font.className} lang="en" suppressHydrationWarning>
      <body>
        <SessionProvider session={session}>
          <Providers>
            {children}
            <Toaster />
          </Providers>
          {/* <ProgressBar /> */}
        </SessionProvider>
        <OnlineStatusMonitor />
      </body>
    </html>
  );
}
