import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getTeamObjectivesPage } from "@/modules/objectives/queries/get-team-objectives";
import { resolvePaginationInput } from "../tool-helpers";

export const listTeamObjectivesTool = tool({
  description:
    "List objectives from a specific team. Works for all user roles including guests, as long as they belong to the specified team.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to get objectives from (required)"),
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
    { teamId, searchQuery, page, pageSize },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access team objectives",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };
      const pagination = resolvePaginationInput({ page, pageSize });

      const response = await getTeamObjectivesPage(
        teamId,
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
        message: `Found ${response.objectives.length} objectives for team.`,
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
