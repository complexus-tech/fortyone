"use client";
import posthog from "posthog-js";
import { PostHogProvider as PostHogProviderPrimitive } from "posthog-js/react";
import type { ReactNode } from "react";

if (typeof window !== "undefined") {
  posthog.init(process.env.NEXT_PUBLIC_POSTHOG_KEY!, {
    api_host: process.env.NEXT_PUBLIC_POSTHOG_HOST,
    person_profiles: "identified_only", // or 'always' to create profiles for anonymous users as well
    capture_pageview: false,
    capture_pageleave: true,
  });
}
export const PostHogProvider = ({ children }: { children: ReactNode }) => {
  return (
    <PostHogProviderPrimitive client={posthog}>
      {children}
    </PostHogProviderPrimitive>
  );
};
