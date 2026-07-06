import { tool } from "ai";
import { z } from "zod";
import { auth } from "@/auth";
import { getActivities } from "@/lib/queries/activities/get-activities";
import { filterActivityTimeline, paginateRecords } from "./tool-helpers";

export const activitySummaryTool = tool({
  description:
    "List recent workspace story activities. Use this for questions like what changed this week, who changed priority or estimate, and recent status changes.",
  inputSchema: z.object({
    startDate: z
      .string()
      .optional()
      .describe("Start date for the activity window, ISO date."),
    endDate: z
      .string()
      .optional()
      .describe("End date for the activity window, ISO date."),
    userId: z.string().optional().describe("Filter by activity actor ID."),
    fields: z
      .array(z.string())
      .optional()
      .describe(
        "Filter by changed fields such as status, priority, estimate, assignee, sprint, objective, or labels.",
      ),
    page: z.number().min(1).optional().describe("Page number. Default 1."),
    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Activities per page. Default 20, max 100."),
  }),
  execute: async (
    { startDate, endDate, userId, fields, page, pageSize },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access activities",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug?: string })
        .workspaceSlug;

      if (!workspaceSlug) {
        return { success: false, error: "Workspace context is required" };
      }

      const ctx = { session, workspaceSlug };
      const activities = await getActivities(ctx, { startDate, endDate });
      const filtered = filterActivityTimeline(activities, {
        userId,
        fields,
        since: startDate,
        until: endDate,
      });
      const result = paginateRecords(filtered, { page, pageSize });

      return {
        success: true,
        activities: result.records,
        count: result.records.length,
        pagination: result.pagination,
        filters: { startDate, endDate, userId, fields },
        message: `Found ${result.records.length} recent activit${result.records.length === 1 ? "y" : "ies"}.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to list activity summary",
      };
    }
  },
});
