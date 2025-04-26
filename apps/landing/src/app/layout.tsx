import type { Metadata } from "next";
import { Suspense, type ReactNode } from "react";
import { cn } from "lib";
import { SessionProvider } from "next-auth/react";
import { instrumentSans, satoshi } from "@/styles/fonts";
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
    "Streamline project delivery with Complexus - the all-in-one platform combining powerful OKR tracking, sprint planning, and team collaboration tools. Drive measurable results and team alignment.",
  keywords: [
    "OKR management",
    "project management",
    "team collaboration",
    "sprint planning",
    "objective tracking",
    "task management",
    "agile project management",
    "team productivity",
    "strategic planning",
    "project tracking",
    "kanban boards",
    "team alignment",
    "goal tracking",
    "project roadmap",
    "team objectives",
    "project tracking",
    "project software",
    "project management software",
    "project management software for teams",
  ],
  openGraph: {
    type: "website",
    locale: "en_US",
    title: "Project Management & OKR Software for Teams | Complexus",
    description:
      "Streamline project delivery with Complexus - the all-in-one platform combining powerful OKR tracking, sprint planning, and team collaboration tools. Drive measurable results and team alignment.",
    siteName: "Complexus",
    url: "https://www.complexus.app",
  },
  twitter: {
    card: "summary_large_image",
    site: "@complexus_app",
    creator: "@complexus_app",
    title: "Project Management Software with OKR Framework | Complexus",
    description:
      "Streamline project delivery with Complexus - the all-in-one platform combining powerful OKR tracking, sprint planning, and team collaboration tools. Drive measurable results and team alignment.",
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
  alternates: {
    canonical: "https://www.complexus.app",
  },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html className="dark" lang="en" suppressHydrationWarning>
      <head>
        <JsonLd />
      </head>
      <body className={cn(satoshi.variable, instrumentSans.variable)}>
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
    </html>
  );
}
