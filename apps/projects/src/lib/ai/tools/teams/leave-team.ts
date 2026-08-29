import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { leaveTeamAction } from "@/modules/teams/actions/leave-team";

export const leaveTeam = tool({
  description: "Leave a team. Users can leave teams they are members of.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to leave (required)"),
  }),

  execute: async (
    { teamId },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to leave teams",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const result = await leaveTeamAction(teamId, workspaceSlug);

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to leave team",
        };
      }

      return {
        success: true,
        message: "Successfully left the team.",
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to leave team",
      };
    }
  },
});
