import { z } from "zod";
import { tool } from "ai";
import { withWorkspacePath } from "@/utils";

export const navigation = tool({
  description:
    "Navigate to different pages in the application using proper entity resolution",
  inputSchema: z.object({
    // Navigation target type - required for all navigation
    targetType: z
      .enum([
        "user-profile",
        "team",
        "sprint",
        "objective",
        "story",
        "my-work",
        "summary",
        "analytics",
        "sprints",
        "notifications",
        "settings",
        "roadmaps",
        "billing"
      ])
      .describe("Type of navigation target"),

    // Entity ID for dynamic routes
    entityId: z
      .string()
      .optional()
      .describe(
        "Entity UUID (required for user-profile, sprint, objective, story)",
      ),

    // Team ID for team-scoped routes
    teamId: z
      .string()
      .optional()
      .describe("Team UUID (required for team, sprint, objective)"),

    // Route within team context
    route: z
      .enum([
        "stories",
        "sprints",
        "objectives",
        "backlog",
        "deleted",
        "archived",
      ])
      .optional()
      .describe(
        "Specific route within team context (for team targetType) Only admins can access and navigate to deleted and archived, if a non admin navigates via direct link or this tool, they will be redirected to the stories page",
      ),
  }),

  execute: async ({ targetType, entityId, teamId, route }, { experimental_context }) => {
    let routePath: string;
    let message: string;

    const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;



    switch (targetType) {
      case "my-work":
        routePath = withWorkspacePath("/my-work", workspaceSlug);
        message = "Navigating to my work";
        break;

      case "summary":
        routePath = withWorkspacePath("/summary", workspaceSlug);
        message = "Navigating to summary";
        break;

      case "analytics":
        routePath = withWorkspacePath("/analytics", workspaceSlug);
        message = "Navigating to analytics";
        break;

      case "sprints":
        routePath = withWorkspacePath("/sprints", workspaceSlug);
        message = "Navigating to sprints";
        break;

      case "notifications":
        routePath = withWorkspacePath("/notifications", workspaceSlug);
        message = "Navigating to notifications";
        break;

      case "settings":
        routePath = withWorkspacePath("/settings", workspaceSlug);
        message = "Navigating to settings";
        break;

      case "billing":
        routePath = withWorkspacePath("/settings/workspace/billing", workspaceSlug);
        message = "Navigating to billing";
        break;

      case "roadmaps":
        routePath = withWorkspacePath("/roadmaps", workspaceSlug);
        message = "Navigating to roadmaps";
        break;

      case "user-profile":
        if (!entityId) {
          return {
            error: "Entity ID is required for user profile navigation",
          };
        }
        routePath = withWorkspacePath(`/profile/${entityId}`, workspaceSlug);
        message = "Navigating to user profile";
        break;

      case "story":
        if (!entityId) {
          return {
            error: "Entity ID is required for story navigation",
          };
        }
        routePath = withWorkspacePath(`/story/${entityId}`, workspaceSlug);
        message = "Navigating to story details";
        break;

      // Team-scoped routes requiring both teamId and entityId
      case "sprint":
        if (!teamId || !entityId) {
          return {
            error: "Team ID and Entity ID are required for sprint navigation",
          };
        }
        routePath = withWorkspacePath(`/teams/${teamId}/sprints/${entityId}/stories`, workspaceSlug);
        message = "Navigating to sprint details";
        break;

      case "objective":
        if (!teamId || !entityId) {
          return {
            error:
              "Team ID and Entity ID are required for objective navigation",
          };
        }
        routePath = withWorkspacePath(`/teams/${teamId}/objectives/${entityId}`, workspaceSlug);
        message = "Navigating to objective details";
        break;

      // Team routes requiring only teamId
      case "team":
        if (!teamId) {
          return {
            error: "Team ID is required for team navigation",
          };
        }

        if (route) {
          routePath = withWorkspacePath(`/teams/${teamId}/${route}`, workspaceSlug);
          message = `Navigating to team ${route}`;
        } else {
          routePath = withWorkspacePath(`/teams/${teamId}/stories`, workspaceSlug);
          message = "Navigating to team stories";
        }
        break;

      default:
        return {
          error: "Invalid target type specified",
        };
    }

    return {
      route: routePath,
      type: "navigation",
      targetType,
      message,
    };
  },
});
