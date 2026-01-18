import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getTeams } from "@/modules/teams/queries/get-teams";

export const getTeamDetails = tool({
  description:
    "Get detailed information about a specific team including its members and settings.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to get details for (required)"),
  }),

  execute: async ({ teamId }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access team details",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const teams = await getTeams(ctx);
      const team = teams.find((t) => t.id === teamId);

      if (!team) {
        return {
          success: false,
          error: "Team not found or you don't have access to it",
        };
      }

      return {
        success: true,
        team: {
          id: team.id,
          name: team.name,
          code: team.code,
          memberCount: team.memberCount,
          isPrivate: team.isPrivate,
          color: team.color,
        },
        message: `Retrieved details for team "${team.name}".`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to get team details",
      };
    }
  },
});
