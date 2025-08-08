import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { deleteTeamAction } from "@/modules/teams/actions/delete-team";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const deleteTeam = tool({
  description:
    "Delete a team. Only admins can delete teams. This action cannot be undone.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to delete (required)"),
  }),

  execute: async ({ teamId }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to delete teams",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      // Only admins can delete teams
      if (userRole !== "admin") {
        return {
          success: false,
          error: "Only admins can delete teams",
        };
      }

      const result = await deleteTeamAction(teamId);

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to delete team",
        };
      }

      return {
        success: true,
        message: "Team deleted successfully.",
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to delete team",
      };
    }
  },
});
