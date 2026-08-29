import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { GlobalProfilePage } from "@/modules/public-portal/global-profile-page";
import { getFeedbackSetupHref } from "@/modules/public-portal/feedback-setup";
import { getFeedbackProfileActivity } from "@/modules/public-portal/profile-activity";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";
import { getRedirectUrl } from "@/utils";
import { getLoginUrl } from "@/utils/callback-url";

export const metadata: Metadata = {
  title: "Profile - FortyOne",
  robots: {
    follow: false,
    index: false,
  },
};

export default async function ProfileRoute() {
  const session = await auth();
  if (!session) redirect(getLoginUrl("/profile"));

  const [profile, workspaces, initialActivity] = await Promise.all([
    getProfile(),
    getWorkspaces(),
    getFeedbackProfileActivity("feedback"),
  ]);
  const activeWorkspace =
    workspaces.find(
      (workspace) => workspace.id === profile.lastUsedWorkspaceId,
    ) ?? workspaces.at(0);

  return (
    <GlobalProfilePage
      initialActivity={initialActivity}
      profile={profile}
      viewer={{
        canReceiveUpdates: true,
        id: profile.id,
        kind: "account",
        accountHref: "/account",
        appHref: activeWorkspace
          ? getRedirectUrl(workspaces, [], profile.lastUsedWorkspaceId)
          : undefined,
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
