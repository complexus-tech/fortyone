import type { ReactNode } from "react";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getRedirectUrl } from "@/utils";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";

export default async function Layout({
  children,
  searchParams,
}: {
  children: ReactNode;
  searchParams: Promise<{ mobile?: string }>;
}) {
  const params = await searchParams;
  const isMobile = params?.mobile === "true";
  if (isMobile) {
    const callbackUrl = new URL("fortyone://auth-callback");
    callbackUrl.searchParams.set("code", "test");
    redirect(callbackUrl.toString());
  }
  const session = await auth();

  if (session) {
    const [workspaces, profile] = await Promise.all([
      getWorkspaces(session?.token || ""),
      getProfile(session),
    ]);
    redirect(getRedirectUrl(workspaces, [], profile?.lastUsedWorkspaceId));
  }
  return <>{children}</>;
}
