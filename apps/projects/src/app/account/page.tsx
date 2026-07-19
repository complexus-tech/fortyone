import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { AccountPage } from "@/modules/public-portal/account-page";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";
import { getRedirectUrl } from "@/utils";
import { getLoginUrl } from "@/utils/callback-url";

export const metadata: Metadata = {
  title: "Account settings - FortyOne",
  robots: {
    follow: false,
    index: false,
  },
};

export default async function AccountRoute() {
  const session = await auth();

  if (!session) {
    redirect(getLoginUrl("/account"));
  }

  const [profile, workspaces] = await Promise.all([
    getProfile(),
    getWorkspaces(),
  ]);
  const activeWorkspace =
    workspaces.find(
      (workspace) => workspace.id === profile.lastUsedWorkspaceId,
    ) ?? workspaces.at(0);
  const appHref = activeWorkspace
    ? getRedirectUrl(workspaces, [], profile.lastUsedWorkspaceId)
    : undefined;

  return (
    <AccountPage
      profile={profile}
      viewer={{
        accountHref: "/account",
        appHref,
        avatarUrl: profile.avatarUrl,
        email: profile.email,
        name: profile.fullName || profile.username,
      }}
    />
  );
}
