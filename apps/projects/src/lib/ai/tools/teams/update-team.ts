import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { updateTeamAction } from "@/modules/teams/actions/update-team";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const updateTeam = tool({
  description:
    "Update an existing team. Only admins or team creators can update teams.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to update (required)"),
    name: z.string().optional().describe("Updated team name"),
    color: z.string().optional().describe("Updated team color"),
    isPrivate: z.boolean().optional().describe("Updated privacy setting"),
  }),

  execute: async ({ teamId, name, color, isPrivate }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to update teams",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      // Only admins can update teams for now
      if (userRole !== "admin") {
        return {
          success: false,
          error: "Only admins can update teams",
        };
      }

      const updateData = {
        name,
        color,
        isPrivate,
      };

      const result = await updateTeamAction(teamId, updateData);

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to update team",
        };
      }

      return {
        success: true,
        team: result.data,
        message: `Team "${result.data?.name}" updated successfully.`,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "Failed to update team",
      };
    }
  },
});
