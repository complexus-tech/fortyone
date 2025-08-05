import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getSprints } from "@/modules/sprints/queries/get-sprints";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { getTeamSprints } from "@/modules/sprints/queries/get-team-sprints";

export const listSprints = tool({
  description:
    "List all sprints accessible to the user. Returns sprints with their details, team information, and optional statistics.",
  parameters: z.object({
    teamId: z.string().optional().describe("Team ID to filter sprints by team"),
  }),

  execute: async ({ teamId }) => {
    try {
      const session = await auth();
      if (!session) {
        return {
          success: false,
          error: "Authentication required to access sprints",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;
      let sprints = [];
      if (teamId) {
        sprints = await getTeamSprints(teamId, session);
      } else {
        sprints = await getSprints(session);
      }

      return {
        success: true,
        sprints,
        count: sprints.length,
        userRole,
        message: `Found ${sprints.length} sprint${sprints.length !== 1 ? "s" : ""}.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to list sprints",
      };
    }
  },
});
