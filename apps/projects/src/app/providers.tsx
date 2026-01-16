"use client";

import type { ReactNode } from "react";
import { Suspense } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
// import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { ThemeProvider } from "next-themes";
import TimeAgo from "javascript-time-ago";
import en from "javascript-time-ago/locale/en";
import { getQueryClient } from "./get-query-client";
import { PostHogProvider } from "./posthog";
import GoogleOneTap from "./one-tap";
import PostHogPageView from "./posthog-page-view";

TimeAgo.addDefaultLocale(en);

const isProduction = process.env.NODE_ENV === "production";

export const Providers = ({ children }: { children: ReactNode }) => {
  return (
    <QueryClientProvider client={getQueryClient()}>
      <PostHogProvider>
        <ThemeProvider attribute="class" enableSystem>
          {children}
        </ThemeProvider>
        <Suspense>
          {isProduction ? (
            <>
              <GoogleOneTap />
              <PostHogPageView />
            </>
          ) : null}
        </Suspense>
      </PostHogProvider>
      {/* <ReactQueryDevtools  initialIsOpen={false} /> */}
    </QueryClientProvider>
  );
};
