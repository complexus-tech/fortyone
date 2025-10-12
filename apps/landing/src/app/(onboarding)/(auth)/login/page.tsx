import type { Metadata } from "next";
import { AuthLayout } from "@/modules/auth";
import { auth } from "@/auth";
import { getProfile } from "@/lib/queries/profile";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { redirect } from "next/navigation";
import { getRedirectUrl } from "@/utils";

export const metadata: Metadata = {
  title: "Login - FortyOne",
};

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ mobile?: string }>;
}) {
  const params = await searchParams;
  const isMobile = params?.mobile === "true";

  const session = await auth();

  if (session) {
    if (isMobile) {
      redirect("fortyone://auth-callback?code=test");
    }
    const [workspaces, profile] = await Promise.all([
      getWorkspaces(session?.token || ""),
      getProfile(session),
    ]);
    redirect(getRedirectUrl(workspaces, [], profile?.lastUsedWorkspaceId));
  }
  return <AuthLayout page="login" />;
}
