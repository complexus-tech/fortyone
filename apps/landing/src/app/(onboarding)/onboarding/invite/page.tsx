import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { InviteTeam } from "@/modules/onboarding/invite";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { auth } from "@/auth";
import { getTeams } from "@/lib/queries/get-teams";
import type { Workspace } from "@/types";

export const metadata: Metadata = {
  title: "Invite Team - Complexus",
  description: "Invite your team to Complexus",
};

export default async function InvitePage() {
  const session = await auth();
  const workspaces = await getWorkspaces(session!.token);
  const sortedWorkspaces = workspaces.sort(
    (a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
  );
  let activeWorkspace: Workspace | null = null;
  if (sortedWorkspaces.length > 0) {
    activeWorkspace = sortedWorkspaces[0];
  }
  if (!activeWorkspace || activeWorkspace?.userRole !== "admin") {
    return redirect("/onboarding/welcome");
  }
  const teams = await getTeams(activeWorkspace.id);
  return <InviteTeam activeWorkspace={activeWorkspace} teams={teams} />;
}
