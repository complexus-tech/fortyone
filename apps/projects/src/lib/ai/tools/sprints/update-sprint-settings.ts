import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { updateSprintSettingsAction } from "@/modules/teams/actions/update-sprint-settings";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const updateSprintSettings = tool({
  description:
    "Update sprint automation settings for a team. Only admins can update sprint settings for teams they belong to.",
  inputSchema: z.object({
    teamId: z
      .string()
      .describe("Team ID to update sprint settings for (required)"),
    autoCreateSprints: z
      .boolean()
      .optional()
      .describe("Enable/disable automatic sprint creation"),
    upcomingSprintsCount: z
      .number()
      .min(1)
      .max(4)
      .optional()
      .describe("Number of upcoming sprints to create (1-4)"),
    sprintDurationWeeks: z
      .number()
      .min(1)
      .max(4)
      .optional()
      .describe("Duration of each sprint in weeks (1-4)"),
    sprintStartDay: z
      .enum(["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"])
      .optional()
      .describe("Day of the week when sprints start"),
    moveIncompleteStoriesEnabled: z
      .boolean()
      .optional()
      .describe("Automatically move incomplete stories to the next sprint"),
  }),

  execute: async ({
    teamId,
    autoCreateSprints,
    upcomingSprintsCount,
    sprintDurationWeeks,
    sprintStartDay,
    moveIncompleteStoriesEnabled,
  }, { experimental_context }) => {
    try {
      const session = await auth();
      if (!session) {
        return {
          success: false,
          error: "Authentication required to update sprint settings",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;

      if (userRole !== "admin") {
        return {
          success: false,
          error: "Only admins can update sprint settings",
        };
      }

      const updateData = {
        autoCreateSprints,
        upcomingSprintsCount,
        sprintDurationWeeks,
        sprintStartDay,
        moveIncompleteStoriesEnabled,
      };

      const result = await updateSprintSettingsAction(teamId, updateData, workspaceSlug);

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to update sprint settings",
        };
      }

      return {
        success: true,
        settings: result.data,
        message: "Sprint settings updated.",
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to update sprint settings",
      };
    }
  },
});
