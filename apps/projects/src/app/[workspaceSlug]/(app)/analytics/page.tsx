import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { AnalyticsPage } from "@/modules/analytics";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { withWorkspacePath } from "@/utils";
import { getCookieHeader } from "@/lib/http/header";

export const metadata: Metadata = {
  title: "Analytics",
};

export default async function Page({
  params,
}: {
  params: Promise<{ workspaceSlug: string }>;
}) {
  const [{ workspaceSlug }, session, cookieHeader] = await Promise.all([
    params,
    auth(),
    getCookieHeader(),
  ]);

  if (!session) {
    redirect(withWorkspacePath("/summary", workspaceSlug));
  }

  const workspaces = await getWorkspaces(session.token, cookieHeader);
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === workspaceSlug.toLowerCase(),
  );

  if (workspace?.userRole !== "admin") {
    redirect(withWorkspacePath("/summary", workspaceSlug));
  }

  return <AnalyticsPage />;
}
