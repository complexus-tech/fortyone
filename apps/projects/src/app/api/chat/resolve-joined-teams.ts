import type { Session } from "@/auth";
import type { Team } from "@/modules/teams/types";
import { getJoinedTeams } from "@/modules/teams/queries/get-teams";
import { getChatErrorDiagnostic } from "./chat-errors";

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
    // eslint-disable-next-line no-console -- Diagnostics intentionally omit backend payloads and user content.
    console.error(
      "Failed to resolve joined teams for Maya",
      getChatErrorDiagnostic(error),
    );
    return null;
  }
}
