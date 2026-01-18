"use client";

import { QueryClientProvider } from "@tanstack/react-query";
import { Suspense } from "react";
import { ThemeProvider } from "next-themes";
import { CursorProvider } from "@/context";
import { PostHogProvider } from "@/app/posthog";
import PostHogPageView from "@/app/posthog-page-view";
import { getQueryClient } from "@/app/get-query-client";

const isProduction = process.env.NODE_ENV === "production";

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider attribute="class" enableSystem>
      <QueryClientProvider client={getQueryClient()}>
        <PostHogProvider>
          <CursorProvider>{children}</CursorProvider>
        </PostHogProvider>
        <Suspense>{isProduction ? <PostHogPageView /> : null}</Suspense>
      </QueryClientProvider>
    </ThemeProvider>
  );
}
