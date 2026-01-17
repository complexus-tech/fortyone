import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getRedirectUrl } from "@/utils";
import { getMyInvitations } from "@/lib/queries/get-invitations";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";
import { EmailVerificationCallback } from "./client";

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ mobileApp?: string }>;
}) {
  const params = await searchParams;
  const isMobileApp = params?.mobileApp === "true";
  const session = await auth();
  if (session && !isMobileApp) {
    const [invitations, workspaces, profile] = await Promise.all([
      getMyInvitations(),
      getWorkspaces(session?.token || ""),
      getProfile(session),
    ]);
    redirect(
      getRedirectUrl(
        workspaces,
        invitations.data || [],
        profile?.lastUsedWorkspaceId,
      ),
    );
  }
  return <EmailVerificationCallback />;
}
