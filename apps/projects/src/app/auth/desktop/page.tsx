import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { Text } from "ui";
import { auth } from "@/auth";
import { OnboardingLayout } from "@/components/layouts/onboarding-layout";
import {
  getDesktopAuthPath,
  parseDesktopAuthRequest,
} from "@/lib/desktop-auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getMyInvitationsForCurrentRequest } from "@/modules/invitations/public/server";
import { withCallbackUrl } from "@/utils/callback-url";
import { DesktopAuthorization } from "./client";

export const metadata: Metadata = {
  title: "Sign in to FortyOne desktop",
  robots: { index: false, follow: false },
  referrer: "no-referrer",
};

export default async function DesktopAuthPage({
  searchParams,
}: {
  searchParams: Promise<{
    state?: string | string[];
    code_challenge?: string | string[];
    redirect_uri?: string | string[];
  }>;
}) {
  const transaction = parseDesktopAuthRequest(await searchParams);
  if (!transaction) {
    return (
      <OnboardingLayout>
        <Text>
          This sign-in link is invalid. Return to FortyOne and start again.
        </Text>
      </OnboardingLayout>
    );
  }
  const callbackUrl = getDesktopAuthPath(transaction);
  const session = await auth();
  // Desktop browser authorization retains the regular OAuth and email choices.
  if (!session) redirect(withCallbackUrl("/", callbackUrl));
  const workspaces = await getWorkspaces();
  if (workspaces.length === 0) {
    const invitations = await getMyInvitationsForCurrentRequest();
    const invitation = invitations.find((item) => item.token);
    redirect(
      withCallbackUrl(
        invitation?.token
          ? `/onboarding/join?token=${encodeURIComponent(invitation.token)}`
          : "/onboarding/create",
        callbackUrl,
      ),
    );
  }
  return (
    <OnboardingLayout>
      <DesktopAuthorization
        email={session.user.email}
        transaction={transaction}
      />
    </OnboardingLayout>
  );
}
