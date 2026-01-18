import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getTeamSettings } from "@/modules/teams/queries/get-team-settings";

export const getTeamSettingsTool = tool({
  description:
    "Get team settings including sprint automation and story automation settings. Returns detailed configuration for the team's automation preferences.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to get settings for (required)"),
  }),

  execute: async ({ teamId }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access team settings",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const settings = await getTeamSettings(teamId, ctx);

      return {
        success: true,
        settings,
        message: "Retrieved team settings successfully.",
      };
    } catch (error) {
      return {
        success: false,
        error:
          "Failed to get team settings. You may not have access to this team.",
      };
    }
  },
});
