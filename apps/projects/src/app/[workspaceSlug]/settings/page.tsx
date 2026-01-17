import type { Metadata } from "next";
import { WorkspaceGeneralSettings } from "@/modules/settings/workspace/general";
import { headers } from "next/headers";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";

export const metadata: Metadata = {
  title: "Settings",
};

export default async function Page({
  params,
}: {
  params: Promise<{ workspaceSlug: string }>;
}) {
  const { workspaceSlug } = await params;
  const session = await auth();

  const workspaces = await getWorkspaces(session!.token);
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === workspaceSlug.toLowerCase(),
  );

  if (workspace?.userRole !== "admin") {
    redirect("/settings/account");
  }

  return <WorkspaceGeneralSettings />;
}
