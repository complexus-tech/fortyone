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
    code: z
      .string()
      .optional()
      .describe(
        "Updated team code, (unique identifier 3 characters). eg WEB, DEV, MKT, PRD",
      ),
    isPrivate: z.boolean().optional().describe("Updated privacy setting"),
  }),

  execute: async ({ teamId, name, color, code, isPrivate }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to update teams",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
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
        code,
        isPrivate,
      };

      const result = await updateTeamAction(teamId, updateData, workspaceSlug);

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
