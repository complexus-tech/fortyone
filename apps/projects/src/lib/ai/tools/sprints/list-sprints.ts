import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getSprintsPage } from "@/modules/sprints/queries/get-sprints";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { getTeamSprintsPage } from "@/modules/sprints/queries/get-team-sprints";
import { resolvePaginationInput } from "../tool-helpers";

export const listSprints = tool({
  description:
    "List all sprints accessible to the user. Returns sprints with their details, team information, and optional statistics.",
  inputSchema: z.object({
    teamId: z.string().optional().describe("Team ID to filter sprints by team"),
    searchQuery: z
      .string()
      .optional()
      .describe("Optional search query for sprint name or goal."),
    page: z.number().min(1).optional().describe("Page number. Default 1."),
    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of sprints per page. Default 20, max 100."),
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
          error: "Authentication required to access sprints",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;
      const pagination = resolvePaginationInput({ page, pageSize });
      const response = teamId
        ? await getTeamSprintsPage(
            teamId,
            ctx,
            searchQuery,
            pagination.page,
            pagination.pageSize,
          )
        : await getSprintsPage(
            ctx,
            searchQuery,
            pagination.page,
            pagination.pageSize,
          );

      return {
        success: true,
        sprints: response.sprints,
        count: response.sprints.length,
        pagination: response.pagination,
        userRole,
        message: `Found ${response.sprints.length} sprint${response.sprints.length !== 1 ? "s" : ""}.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to list sprints",
      };
    }
  },
});
