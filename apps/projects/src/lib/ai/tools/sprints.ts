import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getSprints } from "@/modules/sprints/queries/get-sprints";
import { getRunningSprints } from "@/modules/sprints/queries/get-running-sprints";
import { getTeamSprints } from "@/modules/sprints/queries/get-team-sprints";
import { getSprintDetails } from "@/modules/sprints/queries/get-sprint-details";
import { getSprintAnalytics } from "@/modules/sprints/queries/get-sprint-analytics";
import { createSprintAction } from "@/modules/sprints/actions/create-sprint";
import { updateSprintAction } from "@/modules/sprints/actions/update-sprint";
import { deleteSprintAction } from "@/modules/sprints/actions/delete-sprint";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { getStories } from "@/modules/stories/queries/get-stories";
import { updateStoryAction } from "@/modules/story/actions/update-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const sprintsTool = tool({
  description:
    "Comprehensive sprint management: list, create, update, delete sprints, manage sprint stories, get analytics and burndown data. Uses team IDs directly and supports role-based permissions.",
  parameters: z.object({
    action: z
      .enum([
        "list-sprints",
        "list-running-sprints",
        "list-team-sprints",
        "get-sprint-details",
        "create-sprint",
        "update-sprint",
        "delete-sprint",
        "get-sprint-analytics",
        "list-sprint-stories",
        "add-stories-to-sprint",
        "remove-stories-from-sprint",
        "get-available-stories",
      ])
      .describe("The sprint action to perform"),

    // Team identification
    teamId: z
      .string()
      .optional()
      .describe("Team ID for filtering sprints by team"),

    // Sprint identification
    sprintId: z
      .string()
      .optional()
      .describe("Sprint ID for single sprint operations"),

    sprintIds: z
      .array(z.string())
      .optional()
      .describe("Array of sprint IDs for bulk operations"),

    // Story identification
    storyIds: z
      .array(z.string())
      .optional()
      .describe("Array of story IDs for sprint story operations"),

    // For creating sprints
    sprintData: z
      .object({
        name: z.string().describe("Sprint name"),
        goal: z.string().optional().describe("Sprint goal or description"),
        teamId: z
          .string()
          .optional()
          .describe(
            "Team ID (optional - will auto-select if user has only one team)",
          ),
        objectiveId: z
          .string()
          .optional()
          .describe("Objective ID to link sprint"),
        startDate: z.string().describe("Sprint start date (ISO string)"),
        endDate: z.string().describe("Sprint end date (ISO string)"),
      })
      .optional()
      .describe("Sprint data for creation"),

    // For updating sprints
    updateData: z
      .object({
        name: z.string().optional().describe("Updated sprint name"),
        goal: z.string().optional().describe("Updated sprint goal"),
        objectiveId: z.string().optional().describe("Updated objective ID"),
        startDate: z.string().optional().describe("Updated start date"),
        endDate: z.string().optional().describe("Updated end date"),
      })
      .optional()
      .describe("Sprint update data"),

    // Filtering and options
    includeStats: z
      .boolean()
      .optional()
      .describe("Include story statistics in sprint responses"),

    includeAnalytics: z
      .boolean()
      .optional()
      .describe("Include analytics data in sprint responses"),
  }),

  execute: async ({
    action,
    teamId,
    sprintId,
    sprintIds: _sprintIds,
    storyIds,
    sprintData,
    updateData,
    includeStats = true,
    includeAnalytics = false,
  }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access sprints",
        };
      }

      // Get user's workspace and role for permissions
      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      switch (action) {
        case "list-sprints": {
          const sprints = await getSprints(session);

          return {
            success: true,
            sprints: sprints.map((sprint) => ({
              id: sprint.id,
              name: sprint.name,
              goal: sprint.goal,
              teamId: sprint.teamId,
              objectiveId: sprint.objectiveId,
              startDate: sprint.startDate,
              endDate: sprint.endDate,
              createdAt: sprint.createdAt,
              updatedAt: sprint.updatedAt,
              ...(includeStats && { stats: sprint.stats }),
            })),
            count: sprints.length,
            message: `Found ${sprints.length} sprints across all teams.`,
          };
        }

        case "list-running-sprints": {
          const runningSprints = await getRunningSprints(session);

          return {
            success: true,
            sprints: runningSprints.map((sprint) => ({
              id: sprint.id,
              name: sprint.name,
              goal: sprint.goal,
              teamId: sprint.teamId,
              objectiveId: sprint.objectiveId,
              startDate: sprint.startDate,
              endDate: sprint.endDate,
              ...(includeStats && { stats: sprint.stats }),
            })),
            count: runningSprints.length,
            message: `Found ${runningSprints.length} currently running sprints.`,
          };
        }

        case "list-team-sprints": {
          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for list-team-sprints action",
            };
          }

          const teamSprints = await getTeamSprints(teamId, session);

          return {
            success: true,
            sprints: teamSprints.map((sprint) => ({
              id: sprint.id,
              name: sprint.name,
              goal: sprint.goal,
              teamId: sprint.teamId,
              objectiveId: sprint.objectiveId,
              startDate: sprint.startDate,
              endDate: sprint.endDate,
              ...(includeStats && { stats: sprint.stats }),
            })),
            count: teamSprints.length,
            message: `Found ${teamSprints.length} sprints for the specified team.`,
          };
        }

        case "get-sprint-details": {
          if (!sprintId) {
            return {
              success: false,
              error: "Sprint ID is required for get-sprint-details action",
            };
          }

          const [sprintDetails, analytics] = await Promise.all([
            getSprintDetails(sprintId, session),
            includeAnalytics ? getSprintAnalytics(sprintId, session) : null,
          ]);

          return {
            success: true,
            sprint: {
              id: sprintDetails.id,
              name: sprintDetails.name,
              goal: sprintDetails.goal,
              teamId: sprintDetails.teamId,
              objectiveId: sprintDetails.objectiveId,
              startDate: sprintDetails.startDate,
              endDate: sprintDetails.endDate,
              createdAt: sprintDetails.createdAt,
              updatedAt: sprintDetails.updatedAt,
              ...(analytics && { analytics }),
            },
            message: `Retrieved details for sprint "${sprintDetails.name}".`,
          };
        }

        case "create-sprint": {
          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot create sprints",
            };
          }

          if (!sprintData) {
            return {
              success: false,
              error: "Sprint data is required to create a sprint",
            };
          }

          // Handle team selection
          let finalTeamId = sprintData.teamId;

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
              finalTeamId = userTeams[0].id;
            } else {
              // Ask user to specify team
              return {
                success: false,
                error: `Please specify which team to create the sprint for. Available teams: ${userTeams.map((t) => t.name).join(", ")}`,
              };
            }
          }

          const result = await createSprintAction({
            name: sprintData.name,
            goal: sprintData.goal || "",
            teamId: finalTeamId,
            objectiveId: sprintData.objectiveId || null,
            startDate: sprintData.startDate,
            endDate: sprintData.endDate,
          });

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to create sprint",
            };
          }

          const createdSprint = result.data!;

          return {
            success: true,
            sprint: {
              id: createdSprint.id,
              name: createdSprint.name,
              goal: createdSprint.goal,
              teamId: createdSprint.teamId,
              objectiveId: createdSprint.objectiveId,
              startDate: createdSprint.startDate,
              endDate: createdSprint.endDate,
            },
            message: `Successfully created sprint "${sprintData.name}".`,
          };
        }

        case "update-sprint": {
          if (!sprintId) {
            return {
              success: false,
              error: "Sprint ID is required for update-sprint action",
            };
          }

          if (!updateData) {
            return {
              success: false,
              error: "Update data is required for update-sprint action",
            };
          }

          // Check permissions - users can only update sprints for their teams
          // (unless they're admins)
          if (userRole !== "admin") {
            const sprintDetails = await getSprintDetails(sprintId, session);

            const userTeams = await getTeams(session);
            const hasAccess = userTeams.some(
              (team) => team.id === sprintDetails.teamId,
            );

            if (!hasAccess) {
              return {
                success: false,
                error: "You can only update sprints for teams you belong to",
              };
            }
          }

          const result = await updateSprintAction(sprintId, updateData);

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to update sprint",
            };
          }

          return {
            success: true,
            message: "Successfully updated sprint.",
          };
        }

        case "delete-sprint": {
          if (userRole !== "admin") {
            return {
              success: false,
              error: "Only admins can delete sprints",
            };
          }

          if (!sprintId) {
            return {
              success: false,
              error: "Sprint ID is required for delete-sprint action",
            };
          }

          const sprintDetails = await getSprintDetails(sprintId, session);

          const result = await deleteSprintAction(sprintId);

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to delete sprint",
            };
          }

          return {
            success: true,
            message: `Successfully deleted sprint "${sprintDetails.name}".`,
          };
        }

        case "get-sprint-analytics": {
          if (!sprintId) {
            return {
              success: false,
              error: "Sprint ID is required for get-sprint-analytics action",
            };
          }

          const analytics = await getSprintAnalytics(sprintId, session);

          return {
            success: true,
            analytics: {
              sprintId: analytics.sprintId,
              overview: analytics.overview,
              storyBreakdown: analytics.storyBreakdown,
              burndown: analytics.burndown,
              teamAllocation: analytics.teamAllocation,
            },
            message: "Retrieved sprint analytics data.",
          };
        }

        case "list-sprint-stories": {
          if (!sprintId) {
            return {
              success: false,
              error: "Sprint ID is required for list-sprint-stories action",
            };
          }

          const stories = await getStories(session, { sprintId });

          return {
            success: true,
            stories: stories.map((story) => ({
              id: story.id,
              title: story.title,
              priority: story.priority,
              statusId: story.statusId,
              assigneeId: story.assigneeId,
              teamId: story.teamId,
              sprintId: story.sprintId,
              createdAt: story.createdAt,
              updatedAt: story.updatedAt,
            })),
            count: stories.length,
            message: `Found ${stories.length} stories in this sprint.`,
          };
        }

        case "add-stories-to-sprint": {
          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot modify sprint stories",
            };
          }

          if (!sprintId) {
            return {
              success: false,
              error: "Sprint ID is required for add-stories-to-sprint action",
            };
          }

          if (!storyIds || storyIds.length === 0) {
            return {
              success: false,
              error: "Story IDs are required for add-stories-to-sprint action",
            };
          }

          // Update each story to add them to the sprint
          const results = await Promise.allSettled(
            storyIds.map((storyId: string) =>
              updateStoryAction(storyId, { sprintId }),
            ),
          );

          const successful = results.filter(
            (result) => result.status === "fulfilled" && !result.value.error,
          ).length;

          const failed = results.length - successful;

          if (failed === 0) {
            return {
              success: true,
              message: `Successfully added ${successful} stories to sprint.`,
            };
          }
          return {
            success: successful > 0,
            message: `Added ${successful} stories to sprint. ${failed} failed to update.`,
          };
        }

        case "remove-stories-from-sprint": {
          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot modify sprint stories",
            };
          }

          if (!storyIds || storyIds.length === 0) {
            return {
              success: false,
              error:
                "Story IDs are required for remove-stories-from-sprint action",
            };
          }

          // Update each story to remove them from sprint
          const results = await Promise.allSettled(
            storyIds.map((storyId: string) =>
              updateStoryAction(storyId, { sprintId: undefined }),
            ),
          );

          const successful = results.filter(
            (result) => result.status === "fulfilled" && !result.value.error,
          ).length;

          const failed = results.length - successful;

          if (failed === 0) {
            return {
              success: true,
              message: `Successfully removed ${successful} stories from sprint.`,
            };
          }
          return {
            success: successful > 0,
            message: `Removed ${successful} stories from sprint. ${failed} failed to update.`,
          };
        }

        case "get-available-stories": {
          if (!teamId) {
            return {
              success: false,
              error: "Team ID is required for get-available-stories action",
            };
          }

          // Get stories for the team that are not assigned to any sprint
          const stories = await getStories(session, {
            teamId,
            sprintId: undefined, // This should get unassigned stories
          });

          return {
            success: true,
            stories: stories.map((story) => ({
              id: story.id,
              title: story.title,
              priority: story.priority,
              statusId: story.statusId,
              assigneeId: story.assigneeId,
              teamId: story.teamId,
              createdAt: story.createdAt,
            })),
            count: stories.length,
            message: `Found ${stories.length} available stories for the specified team that can be added to sprints.`,
          };
        }
      }
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "An unexpected error occurred while performing sprint operation",
      };
    }
  },
});
