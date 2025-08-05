import { z } from "zod";
import { tool } from "ai";
import { addDays, formatISO } from "date-fns";
import { auth } from "@/auth";
import { getGroupedStories } from "@/modules/stories/queries/get-grouped-stories";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import type { GroupedStoryParams } from "@/modules/stories/types";

export const listDueTomorrow = tool({
  description:
    "List stories that are due tomorrow, grouped by status. Returns stories with their details and deadlines.",
  parameters: z.object({
    teamId: z.string().optional().describe("Team ID to filter stories by team"),
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
        storiesPerGroup: z
          .number()
          .min(1)
          .max(100)
          .default(50)
          .optional()
          .describe("Number of stories to return per group (default: 10)"),
        page: z
          .number()
          .min(1)
          .optional()
          .describe("Page number for pagination (default: 1)"),
      })
      .optional()
      .describe("Optional filters for story queries"),
  }),

  execute: async ({ teamId, filters }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access stories",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      // Calculate tomorrow's range
      const tomorrow = formatISO(addDays(new Date(), 1), {
        representation: "date",
      });
      const dayAfterTomorrow = formatISO(addDays(new Date(), 2), {
        representation: "date",
      });

      const params: GroupedStoryParams = {
        teamIds: teamId ? [teamId] : undefined,
        deadlineAfter: tomorrow,
        deadlineBefore: dayAfterTomorrow,
        groupBy: "status",
        categories: ["backlog", "started", "unstarted"],
        ...filters,
      };

      const result = await getGroupedStories(session, params);

      return {
        success: true,
        stories: result.groups.map((group) => group.stories),
        meta: result.meta,
        userRole,
        message: `Found ${result.meta.totalGroups} stories due tomorrow across ${result.groups.length} status groups.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to list stories due tomorrow",
      };
    }
  },
});
