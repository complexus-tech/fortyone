import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getTeams } from "@/modules/teams/queries/get-teams";

export const listTeams = tool({
  description:
    "List all teams that the current user is a member of. Returns team details including member count and privacy settings.",
  inputSchema: z.object({}),

  execute: async ({}, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access teams",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const teams = await getTeams(ctx);

      return {
        success: true,
        teams,
        message: `You are a member of ${teams.length} team${teams.length !== 1 ? "s" : ""}: ${teams.map((t) => t.name).join(", ")}.`,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "Failed to list teams",
      };
    }
  },
});
