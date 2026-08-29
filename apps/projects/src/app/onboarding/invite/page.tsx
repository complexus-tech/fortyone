import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { InviteTeam } from "@/modules/onboarding/invite";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getTeams } from "@/lib/queries/get-teams";
import { getProfile } from "@/lib/queries/profile";
import { withOnboardingCallbackUrl } from "@/modules/onboarding/routing";

export const metadata: Metadata = {
  title: "Invite Team - FortyOne",
  description: "Invite your team to FortyOne",
};

export default async function InvitePage({
  searchParams,
}: {
  searchParams: Promise<{ callbackUrl?: string }>;
}) {
  const [{ callbackUrl }, workspaces, profile] = await Promise.all([
    searchParams,
    getWorkspaces(),
    getProfile(),
  ]);
  const activeWorkspace = workspaces.find(
    (workspace) => workspace.id === profile.lastUsedWorkspaceId,
  );
  if (!activeWorkspace || activeWorkspace.userRole !== "admin") {
    return redirect(
      withOnboardingCallbackUrl("/onboarding/welcome", callbackUrl),
    );
  }
  const teams = await getTeams(activeWorkspace.slug);
  return (
    <InviteTeam
      activeWorkspace={activeWorkspace}
      callbackUrl={callbackUrl}
      teams={teams}
    />
  );
}
