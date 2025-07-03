import { z } from "zod";
import { tool } from "ai";

export const navigation = tool({
  description:
    "Navigate to different pages of the application, including parameterized routes for specific users, teams, sprints, objectives, and stories",
  parameters: z.object({
    // Simple navigation (existing functionality)
    destination: z
      .enum([
        "/my-work",
        "/summary",
        "/analytics",
        "/sprints",
        "/notifications",
        "/settings",
        "/roadmaps",
      ])
      .optional()
      .describe("Simple destination to navigate to"),

    // Parameterized navigation (new functionality)
    targetType: z
      .enum([
        "user-profile",
        "team-page",
        "team-sprints",
        "team-objectives",
        "team-stories",
        "team-backlog",
        "sprint-details",
        "objective-details",
        "story-details",
      ])
      .optional()
      .describe("Type of parameterized navigation target"),

    // Entity IDs for parameterized navigation
    userId: z.string().optional().describe("User ID for profile navigation"),

    teamId: z
      .string()
      .optional()
      .describe("Team ID for team-related navigation"),

    sprintId: z
      .string()
      .optional()
      .describe("Sprint ID for sprint-specific navigation"),

    objectiveId: z
      .string()
      .optional()
      .describe("Objective ID for objective-specific navigation"),

    storyId: z
      .string()
      .optional()
      .describe("Story ID for story-specific navigation"),

    // Additional context
    context: z
      .object({
        targetName: z
          .string()
          .optional()
          .describe("Human-readable name of the target"),
        teamName: z
          .string()
          .optional()
          .describe("Human-readable team name for context"),
      })
      .optional()
      .describe("Additional context for user feedback"),
  }),

  execute: async ({
    destination,
    targetType,
    userId,
    teamId,
    sprintId,
    objectiveId,
    storyId,
    context,
  }) => {
    // Handle simple navigation (backward compatibility)
    if (destination) {
      return {
        route: destination,
        type: "simple",
        message: `Navigating to ${destination}`,
      };
    }

    // Handle parameterized navigation
    if (targetType) {
      let route: string;
      let message: string;

      switch (targetType) {
        case "user-profile":
          if (!userId) {
            return {
              error: "User ID is required for profile navigation",
            };
          }
          route = `/profile/${userId}`;
          message = context?.targetName
            ? `Navigating to ${context.targetName}'s profile`
            : "Navigating to user profile";
          break;

        case "team-page":
          if (!teamId) {
            return {
              error: "Team ID is required for team navigation",
            };
          }
          route = `/teams/${teamId}/stories`;
          message = context?.teamName
            ? `Navigating to ${context.teamName} stories`
            : "Navigating to team stories";
          break;

        case "team-sprints":
          if (!teamId) {
            return {
              error: "Team ID is required for team sprints navigation",
            };
          }
          route = `/teams/${teamId}/sprints`;
          message = context?.teamName
            ? `Navigating to ${context.teamName} sprints`
            : "Navigating to team sprints";
          break;

        case "team-objectives":
          if (!teamId) {
            return {
              error: "Team ID is required for team objectives navigation",
            };
          }
          route = `/teams/${teamId}/objectives`;
          message = context?.teamName
            ? `Navigating to ${context.teamName} objectives`
            : "Navigating to team objectives";
          break;

        case "team-stories":
          if (!teamId) {
            return {
              error: "Team ID is required for team stories navigation",
            };
          }
          route = `/teams/${teamId}/stories`;
          message = context?.teamName
            ? `Navigating to ${context.teamName} stories`
            : "Navigating to team stories";
          break;

        case "team-backlog":
          if (!teamId) {
            return {
              error: "Team ID is required for team backlog navigation",
            };
          }
          route = `/teams/${teamId}/backlog`;
          message = context?.teamName
            ? `Navigating to ${context.teamName} backlog`
            : "Navigating to team backlog";
          break;

        case "sprint-details":
          if (!teamId || !sprintId) {
            return {
              error: "Team ID and Sprint ID are required for sprint navigation",
            };
          }
          route = `/teams/${teamId}/sprints/${sprintId}/stories`;
          message = context?.targetName
            ? `Navigating to ${context.targetName} sprint`
            : "Navigating to sprint details";
          break;

        case "objective-details":
          if (!teamId || !objectiveId) {
            return {
              error:
                "Team ID and Objective ID are required for objective navigation",
            };
          }
          route = `/teams/${teamId}/objectives/${objectiveId}`;
          message = context?.targetName
            ? `Navigating to ${context.targetName} objective`
            : "Navigating to objective details";
          break;

        case "story-details":
          if (!storyId) {
            return {
              error: "Story ID is required for story navigation",
            };
          }
          route = `/story/${storyId}`;
          message = context?.targetName
            ? `Navigating to story: ${context.targetName}`
            : "Navigating to story details";
          break;

        default:
          return {
            error: "Invalid target type specified",
          };
      }

      return {
        route,
        type: "parameterized",
        targetType,
        message,
        context,
      };
    }

    return {
      error: "Either destination or targetType must be specified",
    };
  },
});
