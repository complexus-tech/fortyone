import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { InviteTeam } from "@/modules/onboarding/invite";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { auth } from "@/auth";
import { getTeams } from "@/lib/queries/get-teams";

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
  if (sortedWorkspaces.length === 0) {
    return redirect("/onboarding/welcome");
  }
  const activeWorkspace = sortedWorkspaces[0];
  const teams = await getTeams(activeWorkspace.id);
  return <InviteTeam activeWorkspace={activeWorkspace} teams={teams} />;
}
