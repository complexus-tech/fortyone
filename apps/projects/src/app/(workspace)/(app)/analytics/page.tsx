import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { headers } from "next/headers";
import { AnalyticsPage } from "@/modules/analytics";
import PostHogClient from "@/app/posthog-server";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";

export const metadata: Metadata = {
  title: "Analytics",
};

export default async function Page() {
  const session = await auth();
  const posthog = PostHogClient();
  const headersList = await headers();
  const subdomain = headersList.get("host")!.split(".")[0];

  const workspaces = await getWorkspaces(session?.token || "");
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === subdomain.toLowerCase(),
  );

  const isAnalyticsEnabled = await posthog.isFeatureEnabled(
    "analytics_page",
    session?.user?.email ?? "",
  );

  if (workspace?.userRole !== "admin" || !isAnalyticsEnabled) {
    redirect("/summary");
  }

  return <AnalyticsPage />;
}
