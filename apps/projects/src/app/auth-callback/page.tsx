import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getMyInvitationsForCurrentRequest } from "@/modules/invitations/public/server";
import { isMobileAuthFlow } from "@/lib/mobile-auth";
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
  const isMobileApp =
    params.mobileApp === "true" || isMobileAuthFlow(params.callbackUrl);
  const callbackUrl = params.callbackUrl;

  const session = await auth();

  if (!session) {
    redirect(getLoginUrl(callbackUrl, isMobileApp));
  }

  if (isMobileAuthFlow(callbackUrl)) {
    redirect(callbackUrl!);
  }
  if (isMobileApp) {
    redirect(
      "/?mobileApp=true&error=Please%20restart%20sign-in%20from%20the%20latest%20FortyOne%20app.",
    );
  }
  const [invitations, workspaces, profile] = await Promise.all([
    getMyInvitationsForCurrentRequest(),
    getWorkspaces(),
    getProfile(),
  ]);
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
