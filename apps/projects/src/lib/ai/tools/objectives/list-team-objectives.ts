import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getTeamObjectives } from "@/modules/objectives/queries/get-team-objectives";

export const listTeamObjectivesTool = tool({
  description:
    "List objectives from a specific team. Works for all user roles including guests, as long as they belong to the specified team.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to get objectives from (required)"),
  }),

  execute: async (({ teamId }), { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {


        return {
      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };
          success: false,
          error: "Authentication required to access team objectives",
        };
      }

      const objectives = await getTeamObjectives(teamId, session);

      return {
        success: true,
        objectives: objectives.map((objective) => ({
          id: objective.id,
          name: objective.name,
          description: objective.description,
          teamId: objective.teamId,
          leadUser: objective.leadUser,
          startDate: objective.startDate,
          endDate: objective.endDate,
          statusId: objective.statusId,
          priority: objective.priority,
          health: objective.health,
          createdAt: objective.createdAt,
          updatedAt: objective.updatedAt,
        })),
        count: objectives.length,
        message: `Found ${objectives.length} objectives for team.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to list team objectives",
      };
    }
  },
});
