import { tool } from "ai";
import { z } from "zod";
import { auth } from "@/auth";
import { getObjectivesPage } from "@/modules/objectives/queries/get-objectives";
import { resolvePaginationInput } from "../tool-helpers";

export const listObjectivesTool = tool({
  description:
    "List all objectives accessible to the user. For guests, shows only objectives from teams they belong to. For members and admins, shows all objectives in the workspace.",
  inputSchema: z.object({
    searchQuery: z
      .string()
      .optional()
      .describe("Optional search query for objective name or description."),
    page: z.number().min(1).optional().describe("Page number. Default 1."),
    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of objectives per page. Default 20, max 100."),
  }),

  execute: async (
    { searchQuery, page, pageSize },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access objectives",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };
      const pagination = resolvePaginationInput({ page, pageSize });

      const response = await getObjectivesPage(
        ctx,
        searchQuery,
        pagination.page,
        pagination.pageSize,
      );

      return {
        success: true,
        objectives: response.objectives.map((objective) => ({
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
        pagination: response.pagination,
        count: response.objectives.length,
        message: `Found ${response.objectives.length} objectives in the workspace.`,
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
