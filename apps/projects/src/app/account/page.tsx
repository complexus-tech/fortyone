import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { AccountPage } from "@/modules/public-portal/account-page";
import { getFeedbackSetupHref } from "@/modules/public-portal/feedback-setup";
import {
  getPortalAccountPathBySlug,
  getViewerProfileHrefByPortalSlug,
} from "@/modules/public-portal/utils";
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

const PORTAL_SLUG_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

export default async function AccountRoute({
  searchParams,
}: {
  searchParams: Promise<{ portal?: string | string[] }>;
}) {
  const resolvedSearchParams = await searchParams;
  const portalParam = resolvedSearchParams.portal;
  const portalSlug = Array.isArray(portalParam) ? portalParam[0] : portalParam;
  const validPortalSlug =
    portalSlug && PORTAL_SLUG_PATTERN.test(portalSlug) ? portalSlug : undefined;
  const profileHref = validPortalSlug
    ? getViewerProfileHrefByPortalSlug(validPortalSlug)
    : undefined;
  const accountCallbackUrl = validPortalSlug
    ? getPortalAccountPathBySlug(validPortalSlug)
    : "/account";
  const session = await auth();

  if (!session) {
    redirect(getLoginUrl(accountCallbackUrl));
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
      profileHref={profileHref}
      viewer={{
        id: profile.id,
        accountHref: accountCallbackUrl,
        appHref,
        avatarUrl: profile.avatarUrl,
        email: profile.email,
        feedbackSetupHref: getFeedbackSetupHref(
          workspaces,
          profile.lastUsedWorkspaceId,
        ),
        name: profile.fullName || profile.username,
      }}
    />
  );
}
