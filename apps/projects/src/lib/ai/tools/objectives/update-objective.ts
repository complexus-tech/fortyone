import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { updateObjective } from "@/modules/objectives/actions/update-objective";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const updateObjectiveTool = tool({
  description:
    "Update an existing objective. Only admins, objective creators, or assigned lead users can update objectives.",
  inputSchema: z.object({
    objectiveId: z.string().describe("Objective ID to update (required)"),
    name: z.string().optional().describe("Updated objective name"),
    description: z
      .string()
      .optional()
      .describe("Updated objective description (HTML format)"),
    leadUser: z.string().optional().describe("Updated lead user ID"),
    startDate: z
      .string()
      .optional()
      .describe("Updated start date (ISO  date string e.g 2005-06-13)"),
    endDate: z
      .string()
      .optional()
      .describe("Updated end date (ISO  date string e.g 2005-06-13)"),
    statusId: z.string().optional().describe("Updated status ID"),
    priority: z
      .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
      .optional()
      .describe("Updated priority"),
    health: z
      .enum(["On Track", "At Risk", "Off Track"])
      .optional()
      .describe("Updated health status"),
  }),

  execute: async ({
    objectiveId,
    name,
    description,
    leadUser,
    startDate,
    endDate,
    statusId,
    priority,
    health,
  }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {


        return {
          success: false,
          error: "Authentication required to update objectives",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;
      const userId = session.user!.id;

      // Check if user can update this objective
      const objective = await getObjective(objectiveId, ctx);
      if (!objective) {
        return {
          success: false,
          error: "Objective not found or access denied",
        };
      }

      const canUpdate =
        userRole === "admin" ||
        objective.createdBy === userId ||
        objective.leadUser === userId;

      if (!canUpdate) {
        return {
          success: false,
          error: "You don't have permission to update this objective",
        };
      }

      const result = await updateObjective(objectiveId, {
        name,
        description,
        leadUser,
        startDate,
        endDate,
        statusId,
        priority,
        health,
      }, workspaceSlug);

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to update objective",
        };
      }

      return {
        success: true,
        message: `Successfully updated objective "${objective.name}".`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to update objective",
      };
    }
  },
});
