import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getStoryActivities } from "@/modules/story/queries/get-activities";
import {
  describeBackendPage,
  filterActivityTimeline,
  resolvePaginationInput,
} from "./tool-helpers";

export const storyActivitiesTool = tool({
  description:
    "View paginated story activity and changes: track who made what changes and when, with explicit metadata when more history is available.",
  inputSchema: z.object({
    action: z
      .enum(["list-activities", "get-story-timeline", "get-recent-changes"])
      .describe("The activity operation to perform"),

    storyId: z.string().describe("Story ID for activity operations"),

    activityId: z
      .string()
      .optional()
      .describe("Activity ID for specific activity operations"),

    userId: z
      .string()
      .optional()
      .describe("User ID to filter activities by specific user"),

    field: z
      .enum([
        "title",
        "description",
        "status",
        "priority",
        "assignee",
        "startDate",
        "endDate",
        "objective",
        "sprint",
        "labels",
        "estimate",
      ])
      .optional()
      .describe(
        "Filter activities by specific field changes. The estimate field represents relative complexity, not time.",
      ),
    fields: z
      .array(z.string())
      .optional()
      .describe("Filter activities by one or more field names."),

    includeDetails: z
      .boolean()
      .optional()
      .describe("Include detailed field change information"),

    limit: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Limit number of activities returned (default: 20, max: 100)"),

    since: z
      .string()
      .optional()
      .describe(
        "Filter activities since this date (ISO date string e.g 2005-06-13)",
      ),
    until: z
      .string()
      .optional()
      .describe(
        "Filter activities until this date (ISO date string e.g 2005-06-13)",
      ),
    page: z.number().min(1).optional().describe("Page number. Default 1."),
    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of activities per page. Default 20, max 100."),
  }),

  execute: async (
    {
      action,
      storyId,
      userId,
      field,
      fields,
      limit = 20,
      since,
      until,
      page,
      pageSize,
    },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access story activities",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      switch (action) {
        case "list-activities": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required for listing activities",
            };
          }

          const requestedPage = resolvePaginationInput({
            page,
            pageSize: pageSize ?? limit,
          });
          const response = await getStoryActivities(
            storyId,
            ctx,
            requestedPage.page,
            requestedPage.pageSize,
          );
          const activities = filterActivityTimeline(response.activities, {
            userId,
            fields: fields ?? (field ? [field] : undefined),
            since,
            until,
          });
          const pagination = describeBackendPage(
            response.pagination,
            response.activities.length,
          );

          return {
            success: true,
            activities,
            count: activities.length,
            pagination: {
              ...pagination,
              matchingCount: activities.length,
              filtersAppliedTo: "current-page",
            },
            message: `Returned ${activities.length} matching activit${activities.length === 1 ? "y" : "ies"} from story activity page ${pagination.page}${pagination.hasMore ? "; more pages are available" : ""}.`,
          };
        }

        case "get-story-timeline": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required for getting story timeline",
            };
          }

          const requestedPage = resolvePaginationInput({
            page,
            pageSize: pageSize ?? limit,
          });
          const response = await getStoryActivities(
            storyId,
            ctx,
            requestedPage.page,
            requestedPage.pageSize,
          );
          const activities = filterActivityTimeline(response.activities, {
            userId,
            fields: fields ?? (field ? [field] : undefined),
            since,
            until,
          });

          // Sort by creation date (oldest first for timeline)
          activities.sort(
            (a, b) =>
              new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
          );

          const timeline = activities.map((activity) => ({
            id: activity.id,
            type: activity.type,
            field: activity.field,
            currentValue: activity.currentValue,
            createdAt: activity.createdAt,
            userId: activity.userId,
          }));
          const pagination = describeBackendPage(
            response.pagination,
            response.activities.length,
          );

          return {
            success: true,
            timeline,
            count: timeline.length,
            pagination: {
              ...pagination,
              matchingCount: timeline.length,
              filtersAppliedTo: "current-page",
            },
            summary: {
              changesReturned: timeline.length,
              uniqueUsers: new Set(timeline.map((t) => t.userId)).size,
              dateRange: {
                start: timeline.length > 0 ? timeline[0].createdAt : null,
                end:
                  timeline.length > 0
                    ? timeline[timeline.length - 1].createdAt
                    : null,
              },
            },
            message: `Returned ${timeline.length} matching timeline change${timeline.length === 1 ? "" : "s"} from story activity page ${pagination.page}${pagination.hasMore ? "; more pages are available" : ""}.`,
          };
        }

        case "get-recent-changes": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required for getting recent changes",
            };
          }

          const response = await getStoryActivities(storyId, ctx, 1, limit);
          const activities = filterActivityTimeline(response.activities, {
            userId,
            fields: fields ?? (field ? [field] : undefined),
            since,
            until,
          });

          // Sort by creation date (newest first)
          activities.sort(
            (a, b) =>
              new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
          );

          const recentActivities = activities.slice(0, limit);

          const recentChanges = recentActivities.map((activity) => ({
            id: activity.id,
            type: activity.type,
            field: activity.field,
            currentValue: activity.currentValue,
            createdAt: activity.createdAt,
            userId: activity.userId,
          }));
          const pagination = describeBackendPage(
            response.pagination,
            response.activities.length,
          );

          return {
            success: true,
            recentChanges,
            count: recentChanges.length,
            limit,
            pagination: {
              ...pagination,
              matchingCount: recentChanges.length,
              filtersAppliedTo: "current-page",
            },
            message: `Returned ${recentChanges.length} recent matching change${recentChanges.length === 1 ? "" : "s"} from the latest ${limit}-activity page${pagination.hasMore ? "; older activity is available on later pages" : ""}.`,
          };
        }

        default:
          return {
            success: false,
            error: "Invalid activity action",
          };
      }
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "An error occurred",
      };
    }
  },
});
