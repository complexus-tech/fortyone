import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { createSprintAction } from "@/modules/sprints/actions/create-sprint";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const createSprint = tool({
  description:
    "Create a new sprint. Guests cannot create sprints. Members and admins can create sprints for teams they belong to.",
  inputSchema: z.object({
    name: z.string().describe("Sprint name (required)"),
    goal: z
      .string()
      .optional()
      .describe("Sprint goal or description (HTML format)"),
    teamId: z.string().describe("Team ID where sprint belongs (required)"),
    objectiveId: z.string().optional().describe("Objective ID to link sprint"),
    startDate: z.string().describe("Sprint start date (ISO string) (required)"),
    endDate: z.string().describe("Sprint end date (ISO string) (required)"),
  }),

  execute: async ({ name, goal, teamId, objectiveId, startDate, endDate }) => {
    try {
      const session = await auth();
      if (!session) {
        return {
          success: false,
          error: "Authentication required to create sprints",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      if (userRole === "guest") {
        return {
          success: false,
          error: "Guests can only create sprints for teams they belong to",
        };
      }

      const sprintData = {
        name,
        goal,
        teamId,
        objectiveId,
        startDate,
        endDate,
      };

      const result = await createSprintAction(sprintData);

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to create sprint",
        };
      }

      return {
        success: true,
        sprint: result.data,
        message: `Sprint "${name}" created successfully.`,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to create sprint",
      };
    }
  },
});
