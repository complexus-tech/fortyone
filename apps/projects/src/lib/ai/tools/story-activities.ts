import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getStoryActivities } from "@/modules/story/queries/get-activities";

export const storyActivitiesTool = tool({
  description:
    "View story activity timeline and changes: track who made what changes, when they happened, and provide detailed history of story modifications. Perfect for understanding story evolution and accountability.",
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
      ])
      .optional()
      .describe("Filter activities by specific field changes"),

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
  }),

  execute: async ({ action, storyId, limit = 20 }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access story activities",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      switch (action) {
        case "list-activities": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required for listing activities",
            };
          }

          const response = await getStoryActivities(storyId, ctx);
          const activities = response.activities;

          return {
            success: true,
            activities,
            count: activities.length,
            message: `Found ${activities.length} activity${activities.length !== 1 ? "s" : ""} for this story.`,
          };
        }

        case "get-story-timeline": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required for getting story timeline",
            };
          }

          const response = await getStoryActivities(storyId, ctx);
          const activities = response.activities;

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

          return {
            success: true,
            timeline,
            count: timeline.length,
            summary: {
              totalChanges: timeline.length,
              uniqueUsers: new Set(timeline.map((t) => t.userId)).size,
              dateRange: {
                start: timeline.length > 0 ? timeline[0].createdAt : null,
                end:
                  timeline.length > 0
                    ? timeline[timeline.length - 1].createdAt
                    : null,
              },
            },
            message: `Story timeline shows ${timeline.length} change${timeline.length !== 1 ? "s" : ""} over time.`,
          };
        }

        case "get-recent-changes": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required for getting recent changes",
            };
          }

          const response = await getStoryActivities(storyId, ctx);
          const activities = response.activities;

          // Sort by creation date (newest first)
          activities.sort(
            (a, b) =>
              new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
          );

          // Get recent activities (last 10 by default)
          const recentActivities = activities.slice(0, limit);

          const recentChanges = recentActivities.map((activity) => ({
            id: activity.id,
            type: activity.type,
            field: activity.field,
            currentValue: activity.currentValue,
            createdAt: activity.createdAt,
            userId: activity.userId,
          }));

          return {
            success: true,
            recentChanges,
            count: recentChanges.length,
            message: `Found ${recentChanges.length} recent change${recentChanges.length !== 1 ? "s" : ""} to this story.`,
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
