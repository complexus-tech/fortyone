import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getMyInvitationsForCurrentRequest } from "@/modules/invitations/queries/my-invitations-for-current-request";
import { getAuthCode } from "@/lib/queries/get-auth-code";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";
import { getLoginUrl } from "@/utils/callback-url";
import { ClientPage } from "./client";

export const metadata: Metadata = {
  title: "Auth Callback - FortyOne",
  description:
    "Finalizing authentication and routing you to the right FortyOne workspace.",
};

export default async function AuthCallback({
  searchParams,
}: {
  searchParams: Promise<{ callbackUrl?: string; mobileApp?: string }>;
}) {
  const params = await searchParams;
  const isMobileApp = params.mobileApp === "true";
  const callbackUrl = params.callbackUrl;

  const session = await auth();

  if (!session) {
    redirect(getLoginUrl(callbackUrl));
  }

  const [invitations, workspaces, profile] = await Promise.all([
    getMyInvitationsForCurrentRequest(),
    getWorkspaces(),
    getProfile(),
  ]);

  if (isMobileApp && workspaces.length > 0) {
    const authCodeResponse = await getAuthCode();
    if (authCodeResponse.error || !authCodeResponse.data) {
      redirect("/?mobileApp=true&error=Failed to generate auth code");
    } else {
      redirect(
        `fortyone://login?code=${authCodeResponse.data.code}&email=${authCodeResponse.data.email}`,
      );
    }
  }
  return (
    <ClientPage
      callbackUrl={callbackUrl}
      invitations={invitations}
      profile={profile}
      session={session}
      workspaces={workspaces}
    />
  );
}
