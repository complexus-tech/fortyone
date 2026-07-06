import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getGroupedStories } from "@/modules/stories/queries/get-grouped-stories";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import type { GroupedStoryParams } from "@/modules/stories/types";

export const listTeamStories = tool({
  description:
    "List stories across the workspace or within specific teams, grouped by status, assignee, priority, or not grouped. Supports board-grade filters for manager, project manager, developer, and assignee questions.",
  inputSchema: z.object({
    teamId: z
      .string()
      .optional()
      .describe(
        "Optional team ID to get stories from. Omit for workspace-wide story questions.",
      ),
    filters: z
      .object({
        teamIds: z
          .array(z.string())
          .optional()
          .describe(
            "Filter by one or more team IDs. Ignored when teamId is provided.",
          ),
        statusIds: z
          .array(z.string())
          .optional()
          .describe("Filter by status IDs"),
        assigneeIds: z
          .array(z.string())
          .optional()
          .describe("Filter by assignee IDs"),
        reporterIds: z
          .array(z.string())
          .optional()
          .describe("Filter by reporter or creator IDs"),
        titleContains: z
          .string()
          .optional()
          .describe("Filter stories whose title or content contains this text"),
        priorities: z
          .array(z.string())
          .optional()
          .describe("Filter by priorities"),
        sprintIds: z
          .array(z.string())
          .optional()
          .describe("Filter by sprint IDs"),
        labelIds: z
          .array(z.string())
          .optional()
          .describe("Filter by label IDs"),
        estimateValues: z
          .array(z.number().int())
          .optional()
          .describe(
            "Filter by canonical estimate values for the team's estimation scheme",
          ),
        objectiveId: z.string().optional().describe("Filter by objective ID"),
        epicId: z.string().optional().describe("Filter by epic ID"),
        parentId: z.string().optional().describe("Filter by parent story ID"),
        hasNoAssignee: z
          .boolean()
          .optional()
          .describe("Show only stories with no assignee"),
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
      })
      .optional()
      .describe("Optional filters for story queries"),
    groupBy: z
      .enum(["status", "assignee", "priority", "none"])
      .default("status")
      .describe("Group by status, assignee, or priority"),
    orderBy: z
      .enum(["created", "updated", "deadline", "priority"])
      .optional()
      .describe("Sort field for stories inside each group"),
    orderDirection: z.enum(["asc", "desc"]).optional().describe("Sort order"),
  }),

  execute: async (
    { teamId, filters, groupBy, orderBy, orderDirection },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access stories",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;

      const params: GroupedStoryParams = {
        groupBy,
        orderBy,
        orderDirection,
        teamIds: teamId ? [teamId] : filters?.teamIds,
        ...filters,
      };
      if (teamId) {
        params.teamIds = [teamId];
      }

      const result = await getGroupedStories(ctx, params);

      return {
        success: true,
        stories: result.groups.map((group) => ({
          ...group,
          stories: group.stories,
        })),
        meta: result.meta,
        userRole,
        message: teamId
          ? `Found ${result.meta.totalGroups} story groups in this team.`
          : `Found ${result.meta.totalGroups} story groups in this workspace.`,
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
