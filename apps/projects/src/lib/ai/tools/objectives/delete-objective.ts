import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { deleteObjective } from "@/modules/objectives/actions/delete-objective";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const deleteObjectiveTool = tool({
  description:
    "Delete an objective. Only admins or objective creators can delete objectives. This action cannot be undone.",
  inputSchema: z.object({
    objectiveId: z.string().describe("Objective ID to delete (required)"),
  }),

  execute: async ({ objectiveId }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to delete objectives",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;
      const userId = session.user!.id;

      // Check permissions
      const objective = await getObjective(objectiveId, ctx);
      if (!objective) {
        return {
          success: false,
          error: "Objective not found or access denied",
        };
      }

      const canDelete = userRole === "admin" || objective.createdBy === userId;

      if (!canDelete) {
        return {
          success: false,
          error: "Only admins or objective creators can delete objectives",
        };
      }

      const result = await deleteObjective(objectiveId, workspaceSlug);

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to delete objective",
        };
      }

      return {
        success: true,
        message: `Successfully deleted objective "${objective.name}".`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to delete objective",
      };
    }
  },
});
