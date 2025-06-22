import type { Metadata } from "next";
import { Suspense, type ReactNode } from "react";
import { GoogleAnalytics, GoogleTagManager } from "@next/third-parties/google";
import { cn } from "lib";
import { SessionProvider } from "next-auth/react";
import { ThemeProvider } from "next-themes";
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
    "Complexus is a project management and OKR platform that helps engineering, product, and business teams align goals, track progress, and deliver faster. Try it free.",
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
      "Complexus is a project management and OKR platform that helps engineering, product, and business teams align goals, track progress, and deliver faster. Try it free.",
    siteName: "Complexus",
    url: "/",
  },
  twitter: {
    card: "summary_large_image",
    site: "@complexus_app",
    creator: "@complexus_app",
    title: "Project Management & OKR Software for Teams | Complexus",
    description:
      "Complexus is a project management and OKR platform that helps engineering, product, and business teams align goals, track progress, and deliver faster. Try it free.",
  },
  alternates: {
    canonical: "/",
    languages: {
      "en-US": "https://www.complexus.app",
      "x-default": "https://www.complexus.app",
    },
  },
};
const isProduction = process.env.NODE_ENV === "production";

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <JsonLd />
      </head>
      <body className={cn(instrumentSans.variable)}>
        <ThemeProvider attribute="class" enableSystem>
          <SessionProvider>
            <PostHogProvider>
              <CursorProvider>{children}</CursorProvider>
            </PostHogProvider>
            <Suspense>
              {isProduction ? (
                <>
                  <GoogleOneTap />
                  <PostHogPageView />
                </>
              ) : null}
            </Suspense>
          </SessionProvider>
        </ThemeProvider>
        <Toaster />
      </body>
      {isProduction ? (
        <>
          <GoogleAnalytics
            gaId={process.env.NEXT_PUBLIC_GOOGLE_ANALYTICS_ID!}
          />
          <GoogleTagManager gtmId="AW-684738787" />
        </>
      ) : null}
    </html>
  );
}
