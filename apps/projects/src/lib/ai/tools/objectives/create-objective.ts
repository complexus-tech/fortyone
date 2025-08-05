import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { createObjective } from "@/modules/objectives/actions/create-objective";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const createObjectiveTool = tool({
  description:
    "Create a new objective with optional key results. Guests cannot create objectives. Members and admins can create objectives for teams they belong to.",
  parameters: z.object({
    name: z.string().describe("Objective name (required)"),
    description: z
      .string()
      .optional()
      .describe("Objective description (HTML format)"),
    teamId: z
      .string()
      .describe(
        "Team ID where objective belongs (will auto-select if user has only one team)",
      ),
    leadUser: z
      .string()
      .optional()
      .describe("Lead user ID (UUID) for the objective"),
    startDate: z.string().describe("Start date (ISO string, required)"),
    endDate: z.string().describe("End date (ISO string, required)"),
    priority: z
      .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
      .optional()
      .describe("Objective priority"),
    statusId: z.string().describe("Status ID for the objective"),
    keyResults: z
      .array(
        z.object({
          name: z.string().describe("Key result name"),
          measurementType: z
            .enum(["number", "percentage", "boolean"])
            .describe("How the key result is measured"),
          startValue: z.number().describe("Starting value"),
          targetValue: z.number().describe("Target value"),
        }),
      )
      .optional()
      .describe("Key results to create with the objective"),
  }),

  execute: async ({
    name,
    description,
    teamId,
    leadUser,
    startDate,
    endDate,
    priority,
    statusId,
    keyResults,
  }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to create objectives",
        };
      }

      const workspace = await getWorkspace(session);
      const userRole = workspace.userRole;

      if (userRole === "guest") {
        return {
          success: false,
          error: "Guests cannot create objectives",
        };
      }

      const result = await createObjective({
        name,
        description,
        teamId,
        leadUser,
        startDate,
        endDate,
        priority,
        statusId,
        keyResults:
          keyResults?.map((kr) => ({
            ...kr,
            currentValue: kr.startValue,
          })) || [],
      });

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to create objective",
        };
      }

      return {
        success: true,
        objective: result.data,
        message: `Successfully created objective "${result.data?.name}".`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to create objective",
      };
    }
  },
});
