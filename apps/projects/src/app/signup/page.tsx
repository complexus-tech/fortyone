import type { Metadata } from "next";
import { AuthLayout } from "@/modules/auth";
import { auth } from "@/auth";
import { getProfile } from "@/lib/queries/profile";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getCookieHeader } from "@/lib/http/header";
import { redirect } from "next/navigation";
import { getRedirectUrl } from "@/utils";

export const metadata: Metadata = {
  title: "Signup - FortyOne",
  description:
    "Create your FortyOne account to launch projects quickly, invite teammates, and keep work organized with secure, collaborative tools built for modern teams.",
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

  // Mobile app users always see signup form (even if already logged in on web)
  return <AuthLayout page="signup" />;
}
