import type { Session } from "@/auth";
import type { Team } from "@/modules/teams/types";
import { getJoinedTeams } from "@/modules/teams/queries/get-teams";

export async function resolveJoinedTeams({
  session,
  workspaceSlug,
}: {
  session: Session | null;
  workspaceSlug?: string;
}): Promise<Team[] | null> {
  if (!session || !workspaceSlug) {
    return null;
  }

  try {
    return await getJoinedTeams({ session, workspaceSlug });
  } catch (error) {
    // eslint-disable-next-line no-console -- Preserve diagnostics while Maya falls back to a tool lookup.
    console.error("Failed to resolve joined teams for Maya", error);
    return null;
  }
}
