"use client";

import { QueryClientProvider } from "@tanstack/react-query";
import { Suspense } from "react";
import { ThemeProvider } from "next-themes";
import { PostHogProvider } from "@/app/posthog";
import PostHogPageView from "@/app/posthog-page-view";
import { getQueryClient } from "@/app/get-query-client";

const isProduction = process.env.NODE_ENV === "production";

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
      <QueryClientProvider client={getQueryClient()}>
        <PostHogProvider>{children}</PostHogProvider>
        <Suspense>{isProduction ? <PostHogPageView /> : null}</Suspense>
      </QueryClientProvider>
    </ThemeProvider>
  );
}
