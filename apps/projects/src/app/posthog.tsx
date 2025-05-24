"use client";
import posthog from "posthog-js";
import { PostHogProvider as PostHogProviderPrimitive } from "posthog-js/react";
import type { ReactNode } from "react";
import PostHogPageView from "./posthog-page-view";

if (typeof window !== "undefined") {
  posthog.init(process.env.NEXT_PUBLIC_POSTHOG_KEY!, {
    api_host: "/ingest",
    ui_host: process.env.NEXT_PUBLIC_POSTHOG_HOST,
    person_profiles: "identified_only", // or 'always' to create profiles for anonymous users as well
    capture_pageview: false,
    capture_pageleave: true,
    debug: false,
  });
}
export const PostHogProvider = ({ children }: { children: ReactNode }) => {
  return (
    <PostHogProviderPrimitive client={posthog}>
      {/* Page view tracking should be above the children */}
      <PostHogPageView />
      {children}
    </PostHogProviderPrimitive>
  );
};
