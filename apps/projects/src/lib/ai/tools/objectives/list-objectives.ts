import { tool } from "ai";
import { z } from "zod";
import { auth } from "@/auth";
import { getObjectives } from "@/modules/objectives/queries/get-objectives";

export const listObjectivesTool = tool({
  description:
    "List all objectives accessible to the user. For guests, shows only objectives from teams they belong to. For members and admins, shows all objectives in the workspace.",
  inputSchema: z.object({}),

  execute: async ((), { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {


        return {
      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };
          success: false,
          error: "Authentication required to access objectives",
        };
      }
      const objectives = await getObjectives(session);

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
        message: `Found ${objectives.length} objectives in the workspace.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to list objectives",
      };
    }
  },
});
