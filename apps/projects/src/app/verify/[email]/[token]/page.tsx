import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getRedirectUrl } from "@/utils";
import { getMyInvitationsForCurrentRequest } from "@/modules/invitations/queries/my-invitations-for-current-request";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";
import { EmailVerificationCallback } from "./client";

export const metadata: Metadata = {
  title: "Verify Login - FortyOne",
  description:
    "Verify your secure sign-in link to continue to your FortyOne workspace.",
};

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ callbackUrl?: string; mobileApp?: string }>;
}) {
  const params = await searchParams;
  const isMobileApp = params.mobileApp === "true";
  const callbackUrl = params.callbackUrl;
  const session = await auth();
  if (session && !isMobileApp) {
    const [invitations, workspaces, profile] = await Promise.all([
      getMyInvitationsForCurrentRequest(),
      getWorkspaces(),
      getProfile(),
    ]);
    redirect(
      getRedirectUrl(
        workspaces,
        invitations,
        profile.lastUsedWorkspaceId,
        callbackUrl,
      ),
    );
  }
  return (
    <EmailVerificationCallback
      callbackUrl={callbackUrl}
      isMobileApp={isMobileApp}
    />
  );
}
