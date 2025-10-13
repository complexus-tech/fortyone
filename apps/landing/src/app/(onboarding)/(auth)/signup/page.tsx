import type { Metadata } from "next";
import { AuthLayout } from "@/modules/auth";
import { auth } from "@/auth";
import { getProfile } from "@/lib/queries/profile";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { redirect } from "next/navigation";
import { getRedirectUrl } from "@/utils";

export const metadata: Metadata = {
  title: "Signup - FortyOne",
};

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ mobile?: string }>;
}) {
  const params = await searchParams;
  const isMobile = params?.mobile === "true";
  const session = await auth();

  // Only redirect web users if they're already logged in
  if (session && !isMobile) {
    const [workspaces, profile] = await Promise.all([
      getWorkspaces(session?.token || ""),
      getProfile(session),
    ]);
    redirect(getRedirectUrl(workspaces, [], profile?.lastUsedWorkspaceId));
  }

  // Mobile users always see signup form (even if already logged in on web)
  return <AuthLayout page="signup" />;
}
