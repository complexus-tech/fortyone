import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { AnalyticsPage } from "@/modules/analytics";
import PostHogClient from "@/app/posthog-server";
import { auth } from "@/auth";

export const metadata: Metadata = {
  title: "Analytics",
};

export default async function Page() {
  const posthog = PostHogClient();
  const session = await auth();

  const isAnalyticsEnabled = await posthog.isFeatureEnabled(
    "analytics_page",
    session?.user?.email ?? "",
  );

  if (!isAnalyticsEnabled) {
    redirect("/summary");
  }

  return <AnalyticsPage />;
}
