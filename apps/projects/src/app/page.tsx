import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { OnboardingLayout } from "@/components/layouts/onboarding-layout";
import { AuthLayout } from "@/modules/auth";
import { auth } from "@/auth";
import { getProfile } from "@/lib/queries/profile";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getRedirectUrl } from "@/utils";
import { getCookieHeader } from "@/lib/http/header";

export const metadata: Metadata = {
  title: "Login - FortyOne",
  description:
    "Access your FortyOne workspace securely to continue projects, collaborate with your team, and pick up where you left off—fast, private, reliable.",
};

export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ mobileApp?: string }>;
}) {
  const params = await searchParams;
  const isMobileApp = params?.mobileApp === "true";

  const session = await auth();
  const cookieHeader = await getCookieHeader();

  // Only redirect web users if they're already logged in
  if (session && !isMobileApp) {
    const [workspaces, profile] = await Promise.all([
      getWorkspaces(session?.token || "", cookieHeader),
      getProfile({ token: session?.token, cookieHeader }),
    ]);
    redirect(getRedirectUrl(workspaces, [], profile?.lastUsedWorkspaceId));
  }

  // Mobile app users always see login form (even if already logged in on web)
  return (
    <OnboardingLayout>
      <AuthLayout page="login" />
    </OnboardingLayout>
  );
}
