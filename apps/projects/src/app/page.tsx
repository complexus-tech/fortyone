import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { OnboardingLayout } from "@/components/layouts/onboarding-layout";
import { AuthLayout } from "@/modules/auth";
import { getSignInErrorMessage } from "@/modules/auth/errors";
import { auth } from "@/auth";
import { isMobileAuthFlow } from "@/lib/mobile-auth";
import { getProfile } from "@/lib/queries/profile";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getRedirectUrl } from "@/utils";

export const metadata: Metadata = {
  title: "Login - FortyOne",
  description:
    "Access your FortyOne workspace securely to continue projects, collaborate with your team, and pick up where you left off—fast, private, reliable.",
};

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{
    callbackUrl?: string;
    mobileApp?: string;
    error?: string;
    accountDeleted?: string;
    cleanupPending?: string;
  }>;
}) {
  const params = await searchParams;
  const isMobileApp =
    params.mobileApp === "true" || isMobileAuthFlow(params.callbackUrl);
  const errorMessage = params.error;
  const callbackUrl = params.callbackUrl;

  const session = await auth();

  // Only redirect web users if they're already logged in
  if (session && !isMobileApp && !getSignInErrorMessage(errorMessage)) {
    const [workspaces, profile] = await Promise.all([
      getWorkspaces(),
      getProfile(),
    ]);
    redirect(
      getRedirectUrl(workspaces, [], profile.lastUsedWorkspaceId, callbackUrl),
    );
  }

  // Mobile app users always see login form (even if already logged in on web)
  return (
    <OnboardingLayout>
      <AuthLayout
        accountDeleted={params.accountDeleted === "true"}
        callbackUrl={callbackUrl}
        cleanupPending={params.cleanupPending === "true"}
        errorMessage={errorMessage}
        isMobileApp={isMobileApp}
        page="login"
      />
    </OnboardingLayout>
  );
}
