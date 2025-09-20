"use client";

import type { ReactNode } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { ThemeProvider } from "next-themes";
import TimeAgo from "javascript-time-ago";
import en from "javascript-time-ago/locale/en";
import { AIDevtools } from "ai-sdk-devtools";
import { getQueryClient } from "./get-query-client";
import { PostHogProvider } from "./posthog";

TimeAgo.addDefaultLocale(en);

export const Providers = ({ children }: { children: ReactNode }) => {
  return (
    <QueryClientProvider client={getQueryClient()}>
      <PostHogProvider>
        <ThemeProvider attribute="class" enableSystem>
          {process.env.NODE_ENV === "development" && <AIDevtools />}
          {children}
        </ThemeProvider>
      </PostHogProvider>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
};
