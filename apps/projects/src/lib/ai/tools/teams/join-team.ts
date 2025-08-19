import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { addTeamMemberAction } from "@/modules/teams/actions/add-team-member";

export const joinTeam = tool({
  description:
    "Join a public team using its team UUID. Only works for public teams.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to join (required)"),
  }),

  execute: async ({ teamId }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to join teams",
        };
      }

      const userId = session.user!.id!;

      const result = await addTeamMemberAction(teamId, userId);

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
