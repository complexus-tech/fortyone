"use client";

import { ReactNode } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { ThemeProvider } from "next-themes";
import TimeAgo from "javascript-time-ago";
import en from "javascript-time-ago/locale/en";
import { getQueryClient } from "./get-query-client";
import { PostHogProvider } from "./posthog";

TimeAgo.addDefaultLocale(en);

export const Providers = ({ children }: { children: ReactNode }) => {
  return (
    <QueryClientProvider client={getQueryClient()}>
      <PostHogProvider>
        <ThemeProvider attribute="class" enableSystem>
          {children}
        </ThemeProvider>
      </PostHogProvider>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
};
