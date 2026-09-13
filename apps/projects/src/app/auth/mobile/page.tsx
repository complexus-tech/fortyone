import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { Text } from "ui";
import { auth } from "@/auth";
import { OnboardingLayout } from "@/components/layouts/onboarding-layout";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getMyInvitationsForCurrentRequest } from "@/modules/invitations/public/server";
import { getMobileAuthPath, parseMobileAuthRequest } from "@/lib/mobile-auth";
import { withCallbackUrl } from "@/utils/callback-url";
import { MobileAuthorization } from "./client";

export const metadata: Metadata = {
  title: "Sign in to FortyOne mobile",
  robots: { index: false, follow: false },
  referrer: "no-referrer",
};

export default async function MobileAuthPage({
  searchParams,
}: {
  searchParams: Promise<{
    state?: string | string[];
    code_challenge?: string | string[];
  }>;
}) {
  const transaction = parseMobileAuthRequest(await searchParams);
  if (!transaction) {
    return (
      <OnboardingLayout>
        <Text>
          This sign-in link is invalid. Return to the app and start again.
        </Text>
      </OnboardingLayout>
    );
  }
  const callbackUrl = getMobileAuthPath(transaction);
  const session = await auth();
  if (!session) redirect(withCallbackUrl("/?mobileApp=true", callbackUrl));
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
      <MobileAuthorization
        email={session.user.email}
        transaction={transaction}
      />
    </OnboardingLayout>
  );
}
