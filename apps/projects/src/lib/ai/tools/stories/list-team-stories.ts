import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getGroupedStories } from "@/modules/stories/queries/get-grouped-stories";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import type { GroupedStoryParams } from "@/modules/stories/types";

export const listTeamStories = tool({
  description:
    "List all stories from a specific team, grouped by status. Returns stories with their details, assignments, and metadata.",
  inputSchema: z.object({
    teamId: z.string().describe("Team ID to get stories from (required)"),
    filters: z
      .object({
        statusIds: z
          .array(z.string())
          .optional()
          .describe("Filter by status IDs"),
        assigneeIds: z
          .array(z.string())
          .optional()
          .describe("Filter by assignee IDs"),
        priorities: z
          .array(z.string())
          .optional()
          .describe("Filter by priorities"),
        sprintIds: z
          .array(z.string())
          .optional()
          .describe("Filter by sprint IDs"),
        objectiveId: z.string().optional().describe("Filter by objective ID"),
        assignedToMe: z
          .boolean()
          .optional()
          .describe("Show only stories assigned to me"),
        createdByMe: z
          .boolean()
          .optional()
          .describe("Show only stories created by me"),
        createdAfter: z
          .string()
          .optional()
          .describe(
            "Filter stories created after this date (ISO  date string e.g 2005-06-13)",
          ),
        createdBefore: z
          .string()
          .optional()
          .describe(
            "Filter stories created before this date (ISO  date string e.g 2005-06-13)",
          ),
        updatedAfter: z
          .string()
          .optional()
          .describe(
            "Filter stories updated after this date (ISO  date string e.g 2005-06-13)",
          ),
        updatedBefore: z
          .string()
          .optional()
          .describe(
            "Filter stories updated before this date (ISO  date string e.g 2005-06-13)",
          ),
        deadlineAfter: z
          .string()
          .optional()
          .describe(
            "Filter stories with deadlines after this date (ISO  date string e.g 2005-06-13)",
          ),
        deadlineBefore: z
          .string()
          .optional()
          .describe(
            "Filter stories with deadlines before this date (ISO  date string e.g 2005-06-13)",
          ),
        completedAfter: z
          .string()
          .optional()
          .describe(
            "Filter stories completed after this date (ISO  date string e.g 2005-06-13)",
          ),
        completedBefore: z
          .string()
          .optional()
          .describe(
            "Filter stories completed before this date (ISO  date string e.g 2005-06-13)",
          ),
        includeArchived: z
          .boolean()
          .optional()
          .describe("Include archived stories"),
        includeDeleted: z
          .boolean()
          .optional()
          .describe("Include deleted stories"),
        categories: z
          .array(
            z.enum([
              "backlog",
              "unstarted",
              "started",
              "paused",
              "completed",
              "cancelled",
            ]),
          )
          .optional()
          .describe("Filter stories by status categories"),
        storiesPerGroup: z
          .number()
          .min(1)
          .max(100)
          .default(20)
          .optional()
          .describe("Number of stories to return per group (default: 20)"),
        page: z
          .number()
          .min(1)
          .optional()
          .describe("Page number for pagination (default: 1)"),
      })
      .optional()
      .describe("Optional filters for story queries"),
    groupBy: z
      .enum(["status", "assignee", "priority", "none"])
      .default("status")
      .describe("Group by status, assignee, or priority"),
  }),

  execute: async ({ teamId, filters, groupBy }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access stories",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;

      const params: GroupedStoryParams = {
        groupBy,
        teamIds: [teamId],
        ...filters,
      };

      const result = await getGroupedStories(ctx, params);

      return {
        success: true,
        stories: result.groups.map((group) => ({
          ...group,
          stories: group.stories,
        })),
        meta: result.meta,
        userRole,
        message: `Found ${result.meta.totalGroups} status groups in this team.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to list team stories",
      };
    }
  },
});
