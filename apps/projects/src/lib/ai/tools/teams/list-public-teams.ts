import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getPublicTeams } from "@/modules/teams/queries/get-public-teams";
import { getTeams } from "@/modules/teams/queries/get-teams";

export const listPublicTeams = tool({
  description:
    "List all public teams that the current user can join. Returns teams that are not private and that the user is not already a member of.",
  inputSchema: z.object({}),

  execute: async () => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access teams",
        };
      }

      const [publicTeams, userTeams] = await Promise.all([
        getPublicTeams(session),
        getTeams(session),
      ]);

      const userTeamIds = userTeams.map((t) => t.id);
      const availableTeams = publicTeams.filter(
        (team) => !userTeamIds.includes(team.id),
      );

      return {
        success: true,
        teams: availableTeams,
        message: `Found ${availableTeams.length} public team${availableTeams.length !== 1 ? "s" : ""} you can join.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to list public teams",
      };
    }
  },
});
