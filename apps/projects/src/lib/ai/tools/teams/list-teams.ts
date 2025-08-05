import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getTeams } from "@/modules/teams/queries/get-teams";

export const listTeams = tool({
  description:
    "List all teams that the current user is a member of. Returns team details including member count and privacy settings.",
  parameters: z.object({}),

  execute: async () => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access teams",
        };
      }

      const teams = await getTeams(session);

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
