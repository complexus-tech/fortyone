import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getObjective } from "@/modules/objectives/queries/get-objective";

export const getObjectiveDetailsTool = tool({
  description:
    "Get detailed information about a specific objective. Requires appropriate permissions to access the objective. Use objective-statuses tool to get status names.",
  parameters: z.object({
    objectiveId: z
      .string()
      .describe("Objective ID to get details for (required)"),
  }),

  execute: async ({ objectiveId }) => {
    try {
      const session = await auth();
      if (!session) {
        return {
          success: false,
          error: "Authentication required to access objective details",
        };
      }

      const objective = await getObjective(objectiveId, session);

      if (!objective) {
        return {
          success: false,
          error: "Objective not found or access denied",
        };
      }

      return {
        success: true,
        objective: {
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
        },
        message: `Retrieved details for objective "${objective.name}".`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to get objective details",
      };
    }
  },
});
