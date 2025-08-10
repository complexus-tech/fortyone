import type { Metadata } from "next";
import { type ReactNode } from "react";
import { GoogleAnalytics, GoogleTagManager } from "@next/third-parties/google";
import { cn } from "lib";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { body } from "@/styles/fonts";
import "../styles/global.css";
import { JsonLd } from "@/components/shared";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";
import { workspaceKeys } from "@/lib/hooks/workspaces";
import { userKeys } from "@/lib/hooks/profile";
import { Toaster } from "./toaster";
import Providers from "./providers";
import { getQueryClient } from "./get-query-client";

export const metadata: Metadata = {
  title: "Meet Complexus - AI-powered all-in-one Projects & OKRs platform",
  description:
    "Complexus is an AI-powered alternative to Jira, Notion, and Monday built to align teams on Projects & OKRs, track progress, and deliver faster. Try it for free.",
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
    title: "Meet Complexus - AI-powered all-in-one Projects & OKRs platform",
    description:
      "Complexus is an AI-powered alternative to Jira, Notion, and Monday built to align teams on Projects & OKRs, track progress, and deliver faster. Try it for free.",
    siteName: "Complexus",
    url: "/",
  },
  twitter: {
    card: "summary_large_image",
    site: "@complexus_app",
    creator: "@complexus_app",
    title: "Meet Complexus - AI-powered all-in-one Projects & OKRs platform",
    description:
      "Complexus is an AI-powered alternative to Jira, Notion, and Monday built to align teams on Projects & OKRs, track progress, and deliver faster. Try it for free.",
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

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const session = await auth();
  const queryClient = getQueryClient();
  if (session) {
    await Promise.all([
      queryClient.prefetchQuery({
        queryKey: workspaceKeys.lists(),
        queryFn: () => getWorkspaces(session.token),
      }),
      queryClient.prefetchQuery({
        queryKey: userKeys.profile(),
        queryFn: () => getProfile(session),
      }),
    ]);
  }
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <JsonLd />
      </head>
      <body className={cn(body.className)}>
        <Providers session={session}>
          <HydrationBoundary state={dehydrate(queryClient)}>
            {children}
          </HydrationBoundary>
        </Providers>
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
