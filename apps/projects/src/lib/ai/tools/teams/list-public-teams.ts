import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getPublicTeamsPage } from "@/modules/teams/queries/get-public-teams";
import { getTeamsPage } from "@/modules/teams/queries/get-teams";
import { resolvePaginationInput } from "../tool-helpers";

export const listPublicTeams = tool({
  description:
    "List all public teams that the current user can join. Returns teams that are not private and that the user is not already a member of.",
  inputSchema: z.object({
    searchQuery: z
      .string()
      .optional()
      .describe("Optional search query for public team name or code."),
    page: z.number().min(1).optional().describe("Page number. Default 1."),
    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of public teams per page. Default 20, max 100."),
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

      const [publicTeamsPage, userTeamsPage] = await Promise.all([
        getPublicTeamsPage(
          ctx,
          searchQuery,
          pagination.page,
          pagination.pageSize,
        ),
        getTeamsPage(ctx, "", 1, 100),
      ]);

      const userTeamIds = userTeamsPage.teams.map((t) => t.id);
      const availableTeams = publicTeamsPage.teams.filter(
        (team) => !userTeamIds.includes(team.id),
      );

      return {
        success: true,
        teams: availableTeams,
        pagination: publicTeamsPage.pagination,
        message: `Found ${availableTeams.length} public team${availableTeams.length !== 1 ? "s" : ""} you can join.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to list public teams",
      };
    }
  },
});
