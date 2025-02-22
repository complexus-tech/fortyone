import type { Metadata } from "next";
import { Instrument_Sans as InstrumentSans } from "next/font/google";
import { Suspense, type ReactNode } from "react";
import "../styles/global.css";
import { SessionProvider } from "next-auth/react";
import { auth } from "@/auth";
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
  title:
    "Complexus: Empowering Teams to Conquer Project Complexity, Effortlessly.",
  description:
    "Empower your team with innovative project management solutions. Simplify workflows, streamline collaboration, and achieve exceptional results with Complexus. With features like OKR Tracking, Epics Management, Iterations Planning, and Roadmap Visualization, welcome to effortless project management.",
  openGraph: {
    title:
      "Complexus: Empowering Teams to Conquer Project Complexity, Effortlessly.",
    description:
      "Empower your team with innovative project management solutions. Simplify workflows, streamline collaboration, and achieve exceptional results with Complexus. With features like OKR Tracking, Epics Management, Iterations Planning, and Roadmap Visualization, welcome to effortless project management.",
    url: "https://complexus.app",
  },
  twitter: {
    card: "summary_large_image",
    site: "@complexus_tech",
    title:
      "Complexus: Empowering Teams to Conquer Project Complexity, Effortlessly.",
    description:
      "Empower your team with innovative project management solutions. Simplify workflows, streamline collaboration, and achieve exceptional results with Complexus. With features like OKR Tracking, Epics Management, Iterations Planning, and Roadmap Visualization, welcome to effortless project management.",
  },
};

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await auth();
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={font.className}>
        <SessionProvider session={session}>
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
      </body>
    </html>
  );
}
