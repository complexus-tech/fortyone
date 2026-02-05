import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { InviteTeam } from "@/modules/onboarding/invite";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { auth } from "@/auth";
import { getTeams } from "@/lib/queries/get-teams";
import { getProfile } from "@/lib/queries/profile";
import { getCookieHeader } from "@/lib/http/header";

export const metadata: Metadata = {
  title: "Invite Team - FortyOne",
  description: "Invite your team to FortyOne",
};

export default async function InvitePage() {
  const session = await auth();
  const cookieHeader = await getCookieHeader();
  const [workspaces, profile] = await Promise.all([
    getWorkspaces(session?.token, cookieHeader),
    getProfile({ token: session?.token, cookieHeader }),
  ]);
  const activeWorkspace = workspaces.find(
    (workspace) => workspace.id === profile.lastUsedWorkspaceId,
  );
  if (!activeWorkspace || activeWorkspace?.userRole !== "admin") {
    return redirect("/onboarding/welcome");
  }
  const teams = await getTeams(activeWorkspace.slug);
  return <InviteTeam activeWorkspace={activeWorkspace} teams={teams} />;
}
