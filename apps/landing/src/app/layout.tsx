import type { Metadata } from "next";
import { Suspense, type ReactNode } from "react";
import { GoogleAnalytics } from "@next/third-parties/google";
import { cn } from "lib";
import { SessionProvider } from "next-auth/react";
import { instrumentSans } from "@/styles/fonts";
import "../styles/global.css";
import { CursorProvider } from "@/context";
import { JsonLd } from "@/components/shared";
import { PostHogProvider } from "./posthog";
import { Toaster } from "./toaster";
import PostHogPageView from "./posthog-page-view";
import GoogleOneTap from "./one-tap";

export const metadata: Metadata = {
  title: "Project Management & OKR Software for Teams | Complexus",
  description:
    "All-in-one project management & OKR software for teams: streamline sprint planning, task management, and team collaboration with Complexus.",
  metadataBase: new URL("https://www.complexus.app"),
  keywords: [
    "project management",
    "OKR software",
    "team collaboration",
    "sprint planning",
    "task management",
    "agile project management",
    "goal tracking",
    "kanban boards",
    "strategic planning",
    "project roadmap",
    "team objectives",
    "productivity tools",
    "work management",
    "team performance",
    "collaboration platform",
  ],
  openGraph: {
    type: "website",
    locale: "en_US",
    title: "Project Management & OKR Software for Teams | Complexus",
    description:
      "All-in-one project management & OKR software for teams: streamline sprint planning, task management, and team collaboration with Complexus.",
    siteName: "Complexus",
    url: "/",
  },
  twitter: {
    card: "summary_large_image",
    site: "@complexus_app",
    creator: "@complexus_app",
    title: "Project Management & OKR Software for Teams | Complexus",
    description:
      "All-in-one project management & OKR software for teams: streamline sprint planning, task management, and team collaboration with Complexus.",
  },
  alternates: {
    canonical: "/",
    languages: {
      "en-US": "https://www.complexus.app",
      "x-default": "https://www.complexus.app",
    },
  },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html className="dark" lang="en" suppressHydrationWarning>
      <head>
        <JsonLd />
      </head>
      <body className={cn(instrumentSans.variable)}>
        <SessionProvider>
          <PostHogProvider>
            <CursorProvider>{children}</CursorProvider>
          </PostHogProvider>
          <Suspense>
            <GoogleOneTap />
          </Suspense>
          <Suspense>
            <PostHogPageView />
          </Suspense>
        </SessionProvider>
        <Toaster />
      </body>
      {process.env.NEXT_PUBLIC_GOOGLE_ANALYTICS_ID &&
      process.env.NODE_ENV === "production" ? (
        <GoogleAnalytics gaId={process.env.NEXT_PUBLIC_GOOGLE_ANALYTICS_ID} />
      ) : null}
    </html>
  );
}
