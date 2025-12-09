"use client";

import { QueryClientProvider } from "@tanstack/react-query";
import { Suspense } from "react";
import { SessionProvider } from "next-auth/react";
import { ThemeProvider } from "next-themes";
// import type { Session } from "next-auth";
import { CursorProvider } from "@/context";
import { PostHogProvider } from "@/app/posthog";
import PostHogPageView from "@/app/posthog-page-view";
import GoogleOneTap from "@/app/one-tap";
import { getQueryClient } from "@/app/get-query-client";

const isProduction = process.env.NODE_ENV === "production";

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider attribute="class" enableSystem defaultTheme="dark">
      <SessionProvider>
        <QueryClientProvider client={getQueryClient()}>
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
        </QueryClientProvider>
      </SessionProvider>
    </ThemeProvider>
  );
}
