import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { joinPublicTeamAction } from "@/modules/teams/actions/join-public-team";

export const joinTeam = tool({
  description:
    "Join the authenticated user to a public team in the current workspace.",
  inputSchema: z.object({
    teamId: z.uuid().describe("Public team ID to join (required)"),
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
          error: "Authentication required to join teams",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const result = await joinPublicTeamAction(teamId, workspaceSlug);

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to join team",
        };
      }

      return {
        success: true,
        message: "Successfully joined team.",
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to join team",
      };
    }
  },
});
