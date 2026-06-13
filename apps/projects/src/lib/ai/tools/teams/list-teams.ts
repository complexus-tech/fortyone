import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getTeamsPage } from "@/modules/teams/queries/get-teams";
import { resolvePaginationInput } from "../tool-helpers";

export const listTeams = tool({
  description:
    "List all teams that the current user is a member of. Returns team details including member count and privacy settings.",
  inputSchema: z.object({
    searchQuery: z
      .string()
      .optional()
      .describe("Optional search query for team name or code."),
    page: z.number().min(1).optional().describe("Page number. Default 1."),
    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of teams per page. Default 20, max 100."),
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
          error: "Authentication required to access teams",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };
      const pagination = resolvePaginationInput({ page, pageSize });

      const response = await getTeamsPage(
        ctx,
        searchQuery,
        pagination.page,
        pagination.pageSize,
      );

      return {
        success: true,
        teams: response.teams,
        pagination: response.pagination,
        message: `Found ${response.teams.length} team${response.teams.length !== 1 ? "s" : ""}.`,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "Failed to list teams",
      };
    }
  },
});
