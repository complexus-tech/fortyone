import type { ReactNode } from "react";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { withWorkspacePath } from "@/utils";
import { getCookieHeader } from "@/lib/http/header";

export default async function RootLayout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ workspaceSlug: string }>;
}) {
  const { workspaceSlug } = await params;
  const session = await auth();
  const cookieHeader = await getCookieHeader();

  const workspaces = await getWorkspaces(session?.token, cookieHeader);
  const workspace = workspaces.find(
    (w) => w.slug.toLowerCase() === workspaceSlug.toLowerCase(),
  );

  if (workspace?.userRole !== "admin") {
    redirect(withWorkspacePath("/settings/account", workspaceSlug));
  }

  return <>{children}</>;
}
