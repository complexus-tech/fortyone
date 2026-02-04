"use client";

import type { ReactNode } from "react";
import { Suspense, useState } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
// import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { ThemeProvider } from "next-themes";
import TimeAgo from "javascript-time-ago";
import en from "javascript-time-ago/locale/en";
import { NuqsAdapter } from "nuqs/adapters/next/app";
import { useSession } from "next-auth/react";
import { getQueryClient } from "./get-query-client";
import { PostHogProvider } from "./posthog";
import GoogleOneTap from "./one-tap";
import PostHogPageView from "./posthog-page-view";

TimeAgo.addDefaultLocale(en);

const isProduction = process.env.NODE_ENV === "production";

export const Providers = ({ children }: { children: ReactNode }) => {
  const [queryClient] = useState(() => getQueryClient());
  const { status } = useSession();
  const shouldLoadOneTap = isProduction && status === "unauthenticated";

  return (
    <QueryClientProvider client={queryClient}>
      <PostHogProvider>
        <NuqsAdapter>
          <ThemeProvider attribute="class" enableSystem>
            {children}
          </ThemeProvider>
        </NuqsAdapter>
        <Suspense>
          {isProduction ? (
            <>
              {shouldLoadOneTap ? <GoogleOneTap /> : null}
              <PostHogPageView />
            </>
          ) : null}
        </Suspense>
      </PostHogProvider>
      {/* <ReactQueryDevtools  initialIsOpen={false} /> */}
    </QueryClientProvider>
  );
};
