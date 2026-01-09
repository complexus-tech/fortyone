import "ui/styles.css";
import "../styles/global.css";
import type { Metadata, Viewport } from "next";
import {
  Instrument_Sans as InstrumentSans,
  Bricolage_Grotesque as BricolageGrotesque,
} from "next/font/google";
import { type ReactNode } from "react";
import { SessionProvider } from "next-auth/react";
import { cn } from "lib";
import { auth } from "@/auth";
import { Providers } from "./providers";
import { Toaster } from "./toaster";
import { OnlineStatusMonitor } from "./online-monitor";

const font = InstrumentSans({
  subsets: ["latin"],
  variable: "--font-body",
  display: "swap",
  weight: "variable",
});

const heading = BricolageGrotesque({
  variable: "--font-heading",
  display: "swap",
  subsets: ["latin"],
  weight: "variable",
});

export const metadata: Metadata = {
  title: "FortyOne",
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  minimumScale: 1,
  userScalable: false,
};

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await auth();

  return (
    <html
      className={cn(font.variable, heading.variable)}
      lang="en"
      suppressHydrationWarning
    >
      <body>
        <SessionProvider session={session}>
          <Providers>
            {children}
            <span className="text-icon" />
            <Toaster />
          </Providers>
        </SessionProvider>
        <OnlineStatusMonitor />
      </body>
    </html>
  );
}
