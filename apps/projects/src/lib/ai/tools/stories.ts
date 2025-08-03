import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getGroupedStories } from "@/modules/stories/queries/get-grouped-stories";
import type { GroupedStoryParams } from "@/modules/stories/types";
import { getStory } from "@/modules/story/queries/get-story";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { updateStoryAction } from "@/modules/story/actions/update-story";
import { deleteStoryAction } from "@/modules/story/actions/delete-story";
import { bulkUpdateAction } from "@/modules/stories/actions/bulk-update-stories";
import { bulkDeleteAction } from "@/modules/stories/actions/bulk-delete-stories";
import { duplicateStoryAction } from "@/modules/story/actions/duplicate-story";
import { restoreStoryAction } from "@/modules/story/actions/restore-story";
import { getStatuses } from "@/lib/queries/states/get-states";
import { getTeamStatuses } from "@/lib/queries/states/get-team-states";
import { searchQuery } from "@/modules/search/queries/search";
import type { SearchQueryParams } from "@/modules/search/types";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const storiesTool = tool({
  description:
    "Manage stories, view assignments, and handle story operations based on user permissions",
  parameters: z.object({
    action: z
      .enum([
        "list-my-stories",
        "list-created-stories",
        "list-team-stories",
        "search-stories",
        "get-story-details",
        "create-story",
        "update-story",
        "assign-stories-to-user",
        "delete-story",
        "bulk-update-stories",
        "bulk-delete-stories",
        "bulk-create-stories",
        "duplicate-story",
        "restore-story",
        "list-due-soon",
        "list-overdue",
        "list-due-today",
        "list-due-tomorrow",
      ])
      .describe("The story operation to perform"),

    // For filtering and searching
    teamId: z
      .string()
      .optional()
      .describe(
        "Team ID to filter stories - use teams tool to get team ID from name",
      ),
    searchQuery: z
      .string()
      .optional()
      .describe(
        "Full text search query to search story titles and descriptions",
      ),
    // For search-stories action
    statusId: z
      .string()
      .optional()
      .describe(
        "Filter by status ID - use statuses tool to get status ID from name",
      ),
    assigneeId: z
      .string()
      .optional()
      .describe("Filter by single assignee ID (for search only)"),
    priority: z
      .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
      .optional()
      .describe("Filter by single priority (for search only)"),

    // For other list actions
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
          .describe("Filter stories created after this date (ISO string)"),
        createdBefore: z
          .string()
          .optional()
          .describe("Filter stories created before this date (ISO string)"),
        updatedAfter: z
          .string()
          .optional()
          .describe("Filter stories updated after this date (ISO string)"),
        updatedBefore: z
          .string()
          .optional()
          .describe("Filter stories updated before this date (ISO string)"),
        deadlineAfter: z
          .string()
          .optional()
          .describe(
            "Filter stories with deadlines after this date (ISO string)",
          ),
        deadlineBefore: z
          .string()
          .optional()
          .describe(
            "Filter stories with deadlines before this date (ISO string)",
          ),
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
          .describe(
            "Filter stories by status categories (broader than specific status IDs)",
          ),
        storiesPerGroup: z
          .number()
          .min(1)
          .max(100)
          .default(10)
          .optional()
          .describe("Number of stories to return per group (default: 20)"),
        page: z
          .number()
          .min(1)
          .optional()
          .describe("Page number for pagination (default: 1)"),
        limit: z
          .number()
          .min(1)
          .max(100)
          .optional()
          .describe("Maximum number of stories to return (default: 20)"),
      })
      .optional()
      .describe("Optional filters for story queries (not used for search)"),

    // For single story operations
    storyId: z
      .string()
      .optional()
      .describe("Story ID for single story operations"),

    // For bulk operations
    storyIds: z
      .array(z.string())
      .optional()
      .describe("Array of story IDs for bulk operations"),

    // For creating stories
    storyData: z
      .object({
        title: z.string().describe("Story title"),
        description: z.string().optional().describe("Story description"),
        descriptionHTML: z
          .string()
          .optional()
          .describe("Story description HTML"),
        teamId: z
          .string()
          .optional()
          .describe(
            "Team ID where story belongs (optional - will auto-select if user has only one team)",
          ),
        statusId: z.string().optional().describe("Initial status ID"),
        assigneeId: z.string().optional().describe("Assignee user ID"),
        priority: z
          .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
          .optional()
          .describe("Story priority"),
        sprintId: z.string().optional().describe("Sprint ID to assign story"),
        objectiveId: z
          .string()
          .optional()
          .describe("Objective ID to assign story"),
        parentId: z
          .string()
          .optional()
          .describe("Parent story ID for sub-stories"),
        startDate: z
          .string()
          .optional()
          .describe("Story start date (ISO string)"),
        endDate: z.string().optional().describe("Story end date (ISO string)"),
      })
      .optional()
      .describe("Story data for creation"),

    // For bulk creating stories
    storiesData: z
      .array(
        z.object({
          title: z.string().describe("Story title (required)"),
          description: z.string().optional().describe("Story description"),
          descriptionHTML: z
            .string()
            .optional()
            .describe("Story description HTML"),
          teamId: z.string().describe("Team ID where story belongs (required)"),
          statusId: z.string().describe("Initial status ID (required)"),
          assigneeId: z.string().optional().describe("Assignee user ID"),
          priority: z
            .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
            .describe("Story priority (required)"),
          sprintId: z.string().optional().describe("Sprint ID to assign story"),
          objectiveId: z
            .string()
            .optional()
            .describe("Objective ID to assign story"),
          parentId: z
            .string()
            .optional()
            .describe("Parent story ID for sub-stories"),
          startDate: z
            .string()
            .optional()
            .describe("Story start date (ISO string)"),
          endDate: z
            .string()
            .optional()
            .describe("Story end date (ISO string)"),
        }),
      )
      .optional()
      .describe("Array of story data for bulk creation"),

    // For updating stories
    updateData: z
      .object({
        title: z.string().optional().describe("Updated title"),
        description: z.string().optional().describe("Updated description"),
        descriptionHTML: z
          .string()
          .optional()
          .describe("Updated description HTML"),
        statusId: z.string().optional().describe("Updated status ID"),
        assigneeId: z.string().optional().describe("Updated assignee ID"),
        priority: z
          .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
          .optional()
          .describe("Updated priority"),
        sprintId: z.string().optional().describe("Updated sprint ID"),
        objectiveId: z.string().optional().describe("Updated objective ID"),
        startDate: z.string().optional().describe("Updated start date"),
        endDate: z.string().optional().describe("Updated end date"),
      })
      .optional()
      .describe("Story update data"),
  }),

  execute: async ({
    action,
    teamId,
    searchQuery: query,
    statusId,
    assigneeId,
    priority,
    filters,
    storyId,
    storyIds,
    storyData,
    storiesData,
    updateData,
  }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access stories",
        };
      }

      // Get user's workspace and role for permissions
      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;
      const userId = session.user!.id;

      switch (action) {
        case "list-my-stories": {
          const params: GroupedStoryParams = {
            groupBy: "status",
            assignedToMe: true,
            storiesPerGroup: 50,
            ...filters,
          };

          const [groupedResult, allStatuses] = await Promise.all([
            getGroupedStories(session, params),
            getStatuses(session),
          ]);

          // Extract stories from grouped result
          const stories = groupedResult.groups.flatMap(
            (group) => group.stories,
          );

          // Create status lookup map
          const statusMap = new Map(
            allStatuses.map((status) => [status.id, status]),
          );

          return {
            success: true,
            stories: stories.map((story) => {
              const status = statusMap.get(story.statusId);
              return {
                id: story.id,
                title: story.title,
                priority: story.priority,
                statusId: story.statusId,
                status: status
                  ? {
                      name: status.name,
                      color: status.color,
                      category: status.category,
                    }
                  : null,
                assigneeId: story.assigneeId,
                teamId: story.teamId,
                endDate: story.endDate,
                createdAt: story.createdAt,
                updatedAt: story.updatedAt,
              };
            }),
            count: stories.length,
            pagination: {
              totalGroups: groupedResult.meta.totalGroups,
              hasMore: groupedResult.groups.some((group) => group.hasMore),
              groupsWithMore: groupedResult.groups.filter(
                (group) => group.hasMore,
              ).length,
              totalCount: groupedResult.groups.reduce(
                (sum, group) => sum + group.totalCount,
                0,
              ),
              showingCount: stories.length,
            },
            message: `Found ${stories.length} stories assigned to you.`,
          };
        }

        case "list-created-stories": {
          const params: GroupedStoryParams = {
            groupBy: "status",
            createdByMe: true,
            storiesPerGroup: 50,
            ...filters,
          };

          const [groupedResult, allStatuses] = await Promise.all([
            getGroupedStories(session, params),
            getStatuses(session),
          ]);

          // Extract stories from grouped result
          const stories = groupedResult.groups.flatMap(
            (group) => group.stories,
          );

          // Create status lookup map
          const statusMap = new Map(
            allStatuses.map((status) => [status.id, status]),
          );

          return {
            success: true,
            stories: stories.map((story) => {
              const status = statusMap.get(story.statusId);
              return {
                id: story.id,
                title: story.title,
                priority: story.priority,
                statusId: story.statusId,
                status: status
                  ? {
                      name: status.name,
                      color: status.color,
                      category: status.category,
                    }
                  : null,
                assigneeId: story.assigneeId,
                teamId: story.teamId,
                endDate: story.endDate,
                createdAt: story.createdAt,
                updatedAt: story.updatedAt,
              };
            }),
            count: stories.length,
            pagination: {
              totalGroups: groupedResult.meta.totalGroups,
              hasMore: groupedResult.groups.some((group) => group.hasMore),
              groupsWithMore: groupedResult.groups.filter(
                (group) => group.hasMore,
              ).length,
              totalCount: groupedResult.groups.reduce(
                (sum, group) => sum + group.totalCount,
                0,
              ),
              showingCount: stories.length,
            },
            message: `Found ${stories.length} stories created by you.`,
          };
        }

        case "list-team-stories": {
          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for listing team stories",
            };
          }

          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot view team stories",
            };
          }

          const [groupedResult, allStatuses] = await Promise.all([
            getGroupedStories(session, {
              groupBy: "status",
              teamIds: teamId ? [teamId] : undefined,
              storiesPerGroup: 50,
              ...filters,
            }),
            getStatuses(session),
          ]);

          // Extract stories from grouped result
          const stories = groupedResult.groups.flatMap(
            (group) => group.stories,
          );

          // Create status lookup map
          const statusMap = new Map(
            allStatuses.map((status) => [status.id, status]),
          );

          return {
            success: true,
            stories: stories.map((story) => {
              const status = statusMap.get(story.statusId);
              return {
                id: story.id,
                title: story.title,
                priority: story.priority,
                statusId: story.statusId,
                status: status
                  ? {
                      name: status.name,
                      color: status.color,
                      category: status.category,
                    }
                  : null,
                assigneeId: story.assigneeId,
                teamId: story.teamId,
                endDate: story.endDate,
                createdAt: story.createdAt,
                updatedAt: story.updatedAt,
              };
            }),
            count: stories.length,
            pagination: {
              totalGroups: groupedResult.meta.totalGroups,
              hasMore: groupedResult.groups.some((group) => group.hasMore),
              groupsWithMore: groupedResult.groups.filter(
                (group) => group.hasMore,
              ).length,
              totalCount: groupedResult.groups.reduce(
                (sum, group) => sum + group.totalCount,
                0,
              ),
              showingCount: stories.length,
            },
            message: `Found ${stories.length} stories in the team.`,
          };
        }

        case "search-stories": {
          // First get all statuses to convert status name to UUID
          const allStatuses = await getStatuses(session);

          // Use provided status ID directly

          // Build search parameters using single values (as the search API expects)
          const searchParams: SearchQueryParams = {
            query,
            type: "stories",
            ...(teamId && { teamId }),
            ...(statusId && { statusId }),
            ...(assigneeId && { assigneeId }),
            ...(priority && { priority }),
            sortBy: "relevance",
            pageSize: 50,
          };

          const searchResults = await searchQuery(session, searchParams);

          // Create status lookup map
          const statusMap = new Map(
            allStatuses.map((status) => [status.id, status]),
          );

          return {
            success: true,
            stories: searchResults.stories.map((story) => {
              const status = statusMap.get(story.statusId);
              return {
                id: story.id,
                title: story.title,
                priority: story.priority,
                statusId: story.statusId,
                status: status
                  ? {
                      name: status.name,
                      color: status.color,
                      category: status.category,
                    }
                  : null,
                assigneeId: story.assigneeId,
                teamId: story.teamId,
                endDate: story.endDate,
                createdAt: story.createdAt,
                updatedAt: story.updatedAt,
              };
            }),
            count: searchResults.totalStories,
            message: query
              ? `Search for "${query}" returned ${searchResults.totalStories} stories.`
              : `Found ${searchResults.totalStories} stories.`,
          };
        }

        case "get-story-details": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required to get story details",
            };
          }

          const [story, allStatuses] = await Promise.all([
            getStory(storyId, session),
            getStatuses(session),
          ]);

          if (!story) {
            return {
              success: false,
              error: "Story not found or access denied",
            };
          }

          // Find the status for this story
          const status = allStatuses.find((s) => s.id === story.statusId);

          return {
            success: true,
            story: {
              id: story.id,
              title: story.title,
              description: story.description,
              priority: story.priority,
              statusId: story.statusId,
              status: status
                ? {
                    name: status.name,
                    color: status.color,
                    category: status.category,
                  }
                : null,
              assigneeId: story.assigneeId,
              teamId: story.teamId,
              sprintId: story.sprintId,
              objectiveId: story.objectiveId,
              startDate: story.startDate,
              endDate: story.endDate,
              createdAt: story.createdAt,
              updatedAt: story.updatedAt,
              subStories: story.subStories.length || 0,
            },
            message: `Retrieved details for story "${story.title}".`,
          };
        }

        case "create-story": {
          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot create stories",
            };
          }

          if (!storyData) {
            return {
              success: false,
              error: "Story data is required to create a story",
            };
          }

          // Handle team selection
          let finalTeamId = storyData.teamId;

          if (!finalTeamId) {
            // Get user's teams
            const userTeams = await getTeams(session);

            if (userTeams.length === 0) {
              return {
                success: false,
                error:
                  "You don't belong to any teams. Please join a team first.",
              };
            }

            if (userTeams.length === 1) {
              // Auto-select the only team
              finalTeamId = userTeams[0].id;
            } else {
              // Ask user to specify team
              return {
                success: false,
                error: `Please specify which team to create the story in. Available teams: ${userTeams.map((t) => t.name).join(", ")}`,
              };
            }
          }

          // Use provided IDs directly
          const finalAssigneeId = storyData.assigneeId;
          let finalStatusId = storyData.statusId;

          // Use provided assignee ID directly

          // Use provided status ID directly

          // Get default status if none provided
          if (!finalStatusId) {
            const teamStatuses = await getTeamStatuses(finalTeamId, session);

            if (teamStatuses.length === 0) {
              return {
                success: false,
                error:
                  "No statuses found for this team. Please contact an admin.",
              };
            }

            const defaultStatus =
              teamStatuses.find((status) => status.isDefault) ||
              teamStatuses[0];
            finalStatusId = defaultStatus.id;
          }

          const result = await createStoryAction({
            title: storyData.title,
            description: storyData.description || "",
            descriptionHTML: storyData.descriptionHTML || "",
            teamId: finalTeamId,
            statusId: finalStatusId,
            assigneeId: finalAssigneeId || undefined,
            priority: storyData.priority || "No Priority",
            sprintId: storyData.sprintId || undefined,
            objectiveId: storyData.objectiveId || undefined,
            parentId: storyData.parentId || undefined,
            startDate: storyData.startDate || undefined,
            endDate: storyData.endDate || undefined,
            reporterId: userId,
          });

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to create story",
            };
          }

          const createdStory = result.data!;

          // Get the status information for the created story
          const allStatuses = await getStatuses(session);
          const status = allStatuses.find(
            (s) => s.id === createdStory.statusId,
          );

          return {
            success: true,
            story: {
              id: createdStory.id,
              title: createdStory.title,
              teamId: createdStory.teamId,
              priority: createdStory.priority,
              statusId: createdStory.statusId,
              status: status
                ? {
                    name: status.name,
                    color: status.color,
                    category: status.category,
                  }
                : null,
              assigneeId: createdStory.assigneeId,
            },
            message: `Successfully created story "${createdStory.title}".`,
          };
        }

        case "update-story": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required to update a story",
            };
          }

          if (!updateData) {
            return {
              success: false,
              error: "Update data is required to update a story",
            };
          }

          // Check if user can update this story
          const story = await getStory(storyId, session);
          if (!story) {
            return {
              success: false,
              error: "Story not found or access denied",
            };
          }

          const canUpdate =
            userRole === "admin" ||
            story.reporterId === session.user!.id ||
            story.assigneeId === session.user!.id;

          if (!canUpdate) {
            return {
              success: false,
              error: "You don't have permission to update this story",
            };
          }

          const result = await updateStoryAction(storyId, updateData);

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to update story",
            };
          }

          return {
            success: true,
            message: `Successfully updated story "${story.title}".`,
          };
        }

        case "assign-stories-to-user": {
          if (!storyIds || storyIds.length === 0) {
            return {
              success: false,
              error: "Story IDs are required for bulk assignment",
            };
          }

          if (!updateData?.assigneeId) {
            return {
              success: false,
              error: "Assignee ID is required for assignment",
            };
          }

          // Check permissions - users can assign stories to themselves, admins can assign to anyone
          if (userRole !== "admin" && updateData.assigneeId !== userId) {
            return {
              success: false,
              error:
                "You can only assign stories to yourself unless you're an admin",
            };
          }

          const result = await bulkUpdateAction({
            storyIds,
            updates: { assigneeId: updateData.assigneeId },
          });

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to assign stories",
            };
          }

          return {
            success: true,
            message: `Successfully assigned ${storyIds.length} stories.`,
          };
        }

        case "delete-story": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required to delete a story",
            };
          }

          // Check permissions
          const story = await getStory(storyId, session);
          if (!story) {
            return {
              success: false,
              error: "Story not found or access denied",
            };
          }

          const canDelete =
            userRole === "admin" || story.reporterId === session.user?.id;

          if (!canDelete) {
            return {
              success: false,
              error: "Only admins or story creators can delete stories",
            };
          }

          const result = await deleteStoryAction(storyId);

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to delete story",
            };
          }

          const storyTitle = String(story.title);

          return {
            success: true,
            message: `Successfully deleted story "${storyTitle}".`,
          };
        }

        case "bulk-update-stories": {
          if (userRole !== "admin") {
            return {
              success: false,
              error: "Only admins can perform bulk story updates",
            };
          }

          if (!storyIds || storyIds.length === 0) {
            return {
              success: false,
              error: "Story IDs are required for bulk update",
            };
          }

          if (!updateData) {
            return {
              success: false,
              error: "Update data is required for bulk update",
            };
          }

          const result = await bulkUpdateAction({
            storyIds,
            updates: updateData,
          });

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to bulk update stories",
            };
          }

          return {
            success: true,
            message: `Successfully updated ${storyIds.length} stories.`,
          };
        }

        case "bulk-delete-stories": {
          if (userRole !== "admin") {
            return {
              success: false,
              error: "Only admins can perform bulk story deletion",
            };
          }

          if (!storyIds || storyIds.length === 0) {
            return {
              success: false,
              error: "Story IDs are required for bulk delete",
            };
          }

          const result = await bulkDeleteAction(storyIds);

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to bulk delete stories",
            };
          }

          return {
            success: true,
            message: `Successfully deleted ${storyIds.length} stories.`,
          };
        }

        case "bulk-create-stories": {
          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot create stories",
            };
          }

          if (!storiesData) {
            return {
              success: false,
              error: "Stories data is required for bulk creation",
            };
          }

          // Prepare all story creation promises
          const storyPromises = storiesData.map(async (storyData) => {
            return createStoryAction({
              title: storyData.title,
              description: storyData.description || "",
              descriptionHTML: storyData.descriptionHTML || "",
              teamId: storyData.teamId,
              statusId: storyData.statusId,
              assigneeId: storyData.assigneeId || undefined,
              priority: storyData.priority,
              sprintId: storyData.sprintId || undefined,
              objectiveId: storyData.objectiveId || undefined,
              parentId: storyData.parentId || undefined,
              startDate: storyData.startDate || undefined,
              endDate: storyData.endDate || undefined,
              reporterId: userId,
            });
          });

          // Execute all story creations in parallel
          const results = await Promise.all(storyPromises);

          const successful = results
            .filter((result) => result.data)
            .map((result) => result.data);

          const failed = results.filter(
            (result) => result.error?.message,
          ).length;

          return {
            success: successful.length,
            stories: successful,
            failedStories: failed,
            message:
              failed === 0
                ? `Successfully created ${successful.length} stories.`
                : `Created ${successful.length} stories. ${failed} failed to create.`,
          };
        }

        case "duplicate-story": {
          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot duplicate stories",
            };
          }

          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required to duplicate a story",
            };
          }

          const result = await duplicateStoryAction(storyId);

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to duplicate story",
            };
          }

          const duplicatedStory = result.data!;

          return {
            success: true,
            story: {
              id: duplicatedStory.id,
              title: duplicatedStory.title,
            },
            message: `Successfully duplicated story as "${duplicatedStory.title}".`,
          };
        }

        case "restore-story": {
          if (userRole !== "admin") {
            return {
              success: false,
              error: "Only admins can restore deleted stories",
            };
          }

          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required to restore a story",
            };
          }

          const result = await restoreStoryAction(storyId);

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to restore story",
            };
          }

          return {
            success: true,
            message: "Successfully restored the story.",
          };
        }

        case "list-due-today": {
          const today = new Date();
          const todayISO = today.toISOString().split("T")[0];

          const tomorrow = new Date(today);
          tomorrow.setDate(tomorrow.getDate() + 1);
          const tomorrowISO = tomorrow.toISOString().split("T")[0];

          const params: GroupedStoryParams = {
            groupBy: "status",
            assignedToMe: true,
            storiesPerGroup: 50,
            deadlineAfter: todayISO,
            deadlineBefore: tomorrowISO,
            ...filters,
          };

          const [groupedResult, allStatuses] = await Promise.all([
            getGroupedStories(session, params),
            getStatuses(session),
          ]);

          const stories = groupedResult.groups.flatMap(
            (group) => group.stories,
          );
          const statusMap = new Map(
            allStatuses.map((status) => [status.id, status]),
          );

          return {
            success: true,
            stories: stories.map((story) => {
              const status = statusMap.get(story.statusId);
              return {
                id: story.id,
                title: story.title,
                priority: story.priority,
                statusId: story.statusId,
                status: status
                  ? {
                      name: status.name,
                      color: status.color,
                      category: status.category,
                    }
                  : null,
                assigneeId: story.assigneeId,
                teamId: story.teamId,
                endDate: story.endDate,
                createdAt: story.createdAt,
                updatedAt: story.updatedAt,
              };
            }),
            count: stories.length,
            pagination: {
              totalGroups: groupedResult.meta.totalGroups,
              hasMore: groupedResult.groups.some((group) => group.hasMore),
              groupsWithMore: groupedResult.groups.filter(
                (group) => group.hasMore,
              ).length,
              totalCount: groupedResult.groups.reduce(
                (sum, group) => sum + group.totalCount,
                0,
              ),
              showingCount: stories.length,
            },
            message: `Found ${stories.length} stories due today.`,
          };
        }

        case "list-due-soon": {
          const today = new Date();
          const todayISO = today.toISOString().split("T")[0];

          const nextWeek = new Date();
          nextWeek.setDate(nextWeek.getDate() + 7);
          const nextWeekISO = nextWeek.toISOString().split("T")[0];

          const params: GroupedStoryParams = {
            groupBy: "status",
            assignedToMe: true,
            storiesPerGroup: 50,
            deadlineAfter: todayISO,
            deadlineBefore: nextWeekISO,
            ...filters,
          };

          const [groupedResult, allStatuses] = await Promise.all([
            getGroupedStories(session, params),
            getStatuses(session),
          ]);

          const stories = groupedResult.groups.flatMap(
            (group) => group.stories,
          );
          const statusMap = new Map(
            allStatuses.map((status) => [status.id, status]),
          );

          return {
            success: true,
            stories: stories.map((story) => {
              const status = statusMap.get(story.statusId);
              return {
                id: story.id,
                title: story.title,
                priority: story.priority,
                statusId: story.statusId,
                status: status
                  ? {
                      name: status.name,
                      color: status.color,
                      category: status.category,
                    }
                  : null,
                assigneeId: story.assigneeId,
                teamId: story.teamId,
                endDate: story.endDate,
                createdAt: story.createdAt,
                updatedAt: story.updatedAt,
              };
            }),
            count: stories.length,
            pagination: {
              totalGroups: groupedResult.meta.totalGroups,
              hasMore: groupedResult.groups.some((group) => group.hasMore),
              groupsWithMore: groupedResult.groups.filter(
                (group) => group.hasMore,
              ).length,
              totalCount: groupedResult.groups.reduce(
                (sum, group) => sum + group.totalCount,
                0,
              ),
              showingCount: stories.length,
            },
            message: `Found ${stories.length} stories due in the next week.`,
          };
        }

        case "list-overdue": {
          const today = new Date();
          const todayISO = today.toISOString().split("T")[0];

          const params: GroupedStoryParams = {
            groupBy: "status",
            assignedToMe: true,
            storiesPerGroup: 50,
            deadlineBefore: todayISO,
            ...filters,
          };

          const [groupedResult, allStatuses] = await Promise.all([
            getGroupedStories(session, params),
            getStatuses(session),
          ]);

          const stories = groupedResult.groups.flatMap(
            (group) => group.stories,
          );
          const statusMap = new Map(
            allStatuses.map((status) => [status.id, status]),
          );

          return {
            success: true,
            stories: stories.map((story) => {
              const status = statusMap.get(story.statusId);
              return {
                id: story.id,
                title: story.title,
                priority: story.priority,
                statusId: story.statusId,
                status: status
                  ? {
                      name: status.name,
                      color: status.color,
                      category: status.category,
                    }
                  : null,
                assigneeId: story.assigneeId,
                teamId: story.teamId,
                endDate: story.endDate,
                createdAt: story.createdAt,
                updatedAt: story.updatedAt,
              };
            }),
            count: stories.length,
            pagination: {
              totalGroups: groupedResult.meta.totalGroups,
              hasMore: groupedResult.groups.some((group) => group.hasMore),
              groupsWithMore: groupedResult.groups.filter(
                (group) => group.hasMore,
              ).length,
              totalCount: groupedResult.groups.reduce(
                (sum, group) => sum + group.totalCount,
                0,
              ),
              showingCount: stories.length,
            },
            message: `Found ${stories.length} overdue stories.`,
          };
        }

        case "list-due-tomorrow": {
          const tomorrow = new Date();
          tomorrow.setDate(tomorrow.getDate() + 1);
          const tomorrowISO = tomorrow.toISOString().split("T")[0];

          const dayAfterTomorrow = new Date();
          dayAfterTomorrow.setDate(dayAfterTomorrow.getDate() + 2);
          const dayAfterTomorrowISO = dayAfterTomorrow
            .toISOString()
            .split("T")[0];

          const params: GroupedStoryParams = {
            groupBy: "status",
            assignedToMe: true,
            storiesPerGroup: 50,
            deadlineAfter: tomorrowISO,
            deadlineBefore: dayAfterTomorrowISO,
            ...filters,
          };

          const [groupedResult, allStatuses] = await Promise.all([
            getGroupedStories(session, params),
            getStatuses(session),
          ]);

          const stories = groupedResult.groups.flatMap(
            (group) => group.stories,
          );
          const statusMap = new Map(
            allStatuses.map((status) => [status.id, status]),
          );

          return {
            success: true,
            stories: stories.map((story) => {
              const status = statusMap.get(story.statusId);
              return {
                id: story.id,
                title: story.title,
                priority: story.priority,
                statusId: story.statusId,
                status: status
                  ? {
                      name: status.name,
                      color: status.color,
                      category: status.category,
                    }
                  : null,
                assigneeId: story.assigneeId,
                teamId: story.teamId,
                endDate: story.endDate,
                createdAt: story.createdAt,
                updatedAt: story.updatedAt,
              };
            }),
            count: stories.length,
            pagination: {
              totalGroups: groupedResult.meta.totalGroups,
              hasMore: groupedResult.groups.some((group) => group.hasMore),
              groupsWithMore: groupedResult.groups.filter(
                (group) => group.hasMore,
              ).length,
              totalCount: groupedResult.groups.reduce(
                (sum, group) => sum + group.totalCount,
                0,
              ),
              showingCount: stories.length,
            },
            message: `Found ${stories.length} stories due tomorrow.`,
          };
        }
      }
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred while performing story operation",
      };
    }
  },
});
