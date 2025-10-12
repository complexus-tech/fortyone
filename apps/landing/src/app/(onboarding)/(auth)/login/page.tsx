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

export default async function Page() {
  const session = await auth();

  if (session) {
    const [workspaces, profile] = await Promise.all([
      getWorkspaces(session?.token || ""),
      getProfile(session),
    ]);
    redirect(getRedirectUrl(workspaces, [], profile?.lastUsedWorkspaceId));
  }
  return <AuthLayout page="login" />;
}
