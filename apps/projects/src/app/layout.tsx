import type { Metadata, Viewport } from "next";
import { Instrument_Sans as InstrumentSans } from "next/font/google";
import { type ReactNode } from "react";
import "../styles/global.css";
import { SessionProvider } from "next-auth/react";
import { redirect } from "next/navigation";
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
};

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  maximumScale: 1,
  minimumScale: 1,
  userScalable: false,
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#ffffff" },
    { media: "(prefers-color-scheme: dark)", color: "#09090a" },
  ],
};

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await auth();
  if (process.env.NODE_ENV === "production" && !session) {
    redirect("https://www.complexus.app/login");
  }

  if (process.env.NODE_ENV === "development" && !session?.user) {
    redirect("/login");
  }

  return (
    <html className={font.className} lang="en" suppressHydrationWarning>
      <body>
        <SessionProvider session={session}>
          <Providers>
            {children}
            <Toaster />
          </Providers>
        </SessionProvider>
        <OnlineStatusMonitor />
      </body>
    </html>
  );
}
