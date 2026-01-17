import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { headers } from "next/headers";
import { AnalyticsPage } from "@/modules/analytics";
import PostHogClient from "@/app/posthog-server";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { withWorkspacePath } from "@/utils";

export const metadata: Metadata = {
  title: "Analytics",
};

export default async function Page({
  params,
}: {
  params: Promise<{ workspaceSlug: string }>;
}) {
  const { workspaceSlug } = await params;
  const session = await auth();
  const posthog = PostHogClient();

  const workspaces = await getWorkspaces(session?.token || "");
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === workspaceSlug.toLowerCase(),
  );

  const isAnalyticsEnabled = await posthog.isFeatureEnabled(
    "analytics_page",
    session?.user?.email ?? "",
  );

  if (workspace?.userRole !== "admin" || !isAnalyticsEnabled) {
    redirect(withWorkspacePath("/summary", workspaceSlug));
  }

  return <AnalyticsPage />;
}
