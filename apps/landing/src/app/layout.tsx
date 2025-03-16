import type { Metadata } from "next";
import { Suspense, type ReactNode } from "react";
import { cn } from "lib";
import { SessionProvider } from "next-auth/react";
import { instrumentSans, satoshi } from "@/styles/fonts";
import "../styles/global.css";
import { CursorProvider } from "@/context";
import { auth } from "@/auth";
import { JsonLd } from "@/components/shared";
import { PostHogProvider } from "./posthog";
import { Toaster } from "./toaster";
import PostHogPageView from "./posthog-page-view";
import GoogleOneTap from "./one-tap";

export const metadata: Metadata = {
  title: "Complexus | Modern OKR & Project Management Platform",
  description:
    "Transform how teams achieve objectives with Complexus. Powerful OKR tracking, sprint planning, and project management tools that help teams align, execute, and deliver results.",
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
  ],
  openGraph: {
    type: "website",
    locale: "en_US",
    title: "Complexus | Modern OKR & Project Management Platform",
    description:
      "Transform how teams achieve objectives with Complexus. Powerful OKR tracking, sprint planning, and project management tools that help teams align, execute, and deliver results.",
    siteName: "Complexus",
    url: "https://www.complexus.app",
  },
  twitter: {
    card: "summary_large_image",
    site: "@complexus_app",
    creator: "@complexus_app",
    title: "Complexus | Modern OKR & Project Management Platform",
    description:
      "Transform how teams achieve objectives with Complexus. Powerful OKR tracking, sprint planning, and project management tools that help teams align, execute, and deliver results.",
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

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await auth();
  return (
    <html className="dark" lang="en" suppressHydrationWarning>
      <head>
        <JsonLd />
      </head>
      <body
        className={cn(satoshi.variable, instrumentSans.variable, "relative")}
      >
        <SessionProvider session={session}>
          <PostHogProvider>
            <CursorProvider>{children}</CursorProvider>
          </PostHogProvider>
          <GoogleOneTap />
        </SessionProvider>
        <Suspense>
          <PostHogPageView />
        </Suspense>
        <Toaster />
        {/* <div className="pointer-events-none fixed inset-0 z-[3] bg-[url('/noise.png')] bg-repeat opacity-40" /> */}
      </body>
    </html>
  );
}
