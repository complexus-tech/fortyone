import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { searchQuery } from "@/modules/search/queries/search";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import type { SearchQueryParams } from "@/modules/search/types";

export const searchStories = tool({
  description:
    "Search stories using full-text search across titles and descriptions. Supports filtering by team, status, assignee, and priority.",
  inputSchema: z.object({
    searchQuery: z
      .string()
      .describe(
        "Full text search query to search story titles and descriptions (required)",
      ),
    teamId: z.string().optional().describe("Team ID to filter stories by team"),
    statusId: z
      .string()
      .optional()
      .describe(
        "Filter by status ID - use statuses tool to get status ID from name",
      ),
    assigneeId: z.string().optional().describe("Filter by single assignee ID"),
    priority: z
      .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
      .optional()
      .describe("Filter by single priority"),
    limit: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Maximum number of stories to return (default: 20)"),
  }),

  execute: async ({
    searchQuery: query,
    teamId,
    statusId,
    assigneeId,
    priority,
    limit = 20,
  }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to search stories",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      const searchParams: SearchQueryParams = {
        query,
        teamId,
        statusId,
        assigneeId,
        priority,
        pageSize: limit,
        type: "stories",
      };

      const result = await searchQuery(session, searchParams);

      return {
        success: true,
        stories: result.stories,
        totalCount: result.totalStories,
        userRole,
        message: `Found ${result.totalStories} stories matching "${query}".`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to search stories",
      };
    }
  },
});
