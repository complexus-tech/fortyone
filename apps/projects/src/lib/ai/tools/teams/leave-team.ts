import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { removeTeamMemberAction } from "@/modules/teams/actions/remove-team-member";

export const leaveTeam = tool({
  description: "Leave a team. Users can leave teams they are members of.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to leave (required)"),
  }),

  execute: async (({ teamId }), { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {


        return {
      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };
          success: false,
          error: "Authentication required to leave teams",
        };
      }

      const userId = session.user!.id!;

      const result = await removeTeamMemberAction(teamId, userId);

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
