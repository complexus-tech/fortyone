import type { Metadata } from "next";
import { Instrument_Sans as InstrumentSans } from "next/font/google";
import { Suspense, type ReactNode } from "react";
import "../styles/global.css";
import { SessionProvider } from "next-auth/react";
import { SpeedInsights } from "@vercel/speed-insights/next";
import { ProgressBar } from "./progress";
import { Providers } from "./providers";
import { Toaster } from "./toaster";
import PostHogPageView from "./posthog-page-view";
import { OnlineStatusMonitor } from "./online-monitor";

const font = InstrumentSans({
  subsets: ["latin"],
  display: "swap",
  weight: ["400", "500", "600", "700"],
});

export const metadata: Metadata = {
  title: "Complexus",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={font.className}>
        <SessionProvider>
          <Providers>
            {children}
            <Toaster />
          </Providers>
          <Suspense>
            <PostHogPageView />
          </Suspense>
          <ProgressBar />
        </SessionProvider>
        <OnlineStatusMonitor />
        <SpeedInsights />
      </body>
    </html>
  );
}
