import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { createKeyResult } from "@/modules/objectives/actions/create-key-result";
import { getObjective } from "@/modules/objectives/queries/get-objective";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const createKeyResultTool = tool({
  description:
    "Create new key results within objectives. Requires objective ID and key result details.",
  inputSchema: z.object({
    objectiveId: z
      .string()
      .describe("Objective ID where the key result will be created"),
    name: z.string().describe("Key result name"),
    measurementType: z
      .enum(["percentage", "number", "boolean"])
      .describe("How the key result is measured"),
    startValue: z.number().describe("Starting value (0, 1 for boolean)"),
    currentValue: z
      .number()
      .optional()
      .describe(
        "Current progress value (defaults to startValue if not provided)",
      ),
    targetValue: z
      .number()
      .describe("Target value (required for percentage and number types)"),
    startDate: z
      .string()
      .describe("Start date (ISO date string e.g 2005-06-13)"),
    endDate: z.string().describe("End date (ISO date string e.g 2005-06-13)"),
    lead: z.string().describe("Lead user ID (optional)"),
    contributors: z
      .array(z.string())
      .describe("Contributors user IDs (optional)"),
  }),

  execute: async ({
    objectiveId,
    name,
    measurementType,
    startValue,
    currentValue,
    targetValue,
    startDate,
    endDate,
    lead,
    contributors,
  }, { experimental_context }) => {
    const session = await auth();

    if (!session) {


      return {
        success: false,
        error: "Authentication required",
      };
    }

    const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;

    const ctx = { session, workspaceSlug };

    // Get user's workspace and role for permissions
    const workspace = await getWorkspace(ctx);
    const userRole = workspace.userRole;

    if (userRole === "guest") {
      return {
        success: false,
        error: "Guests cannot create key results",
      };
    }

    // Check if user has access to the objective's team
    const objective = await getObjective(objectiveId, ctx);
    if (!objective) {
      return {
        success: false,
        error: "Objective not found",
      };
    }

    const finalCurrentValue = currentValue ?? startValue;
    const result = await createKeyResult({
      objectiveId,
      name,
      measurementType,
      startValue,
      targetValue,
      currentValue: finalCurrentValue,
      startDate,
      endDate,
      lead,
      contributors,
    }, workspaceSlug);

    if (result.error) {
      return {
        success: false,
        error: result.error.message || "Failed to create key result",
      };
    }

    return {
      success: true,
      message: `Successfully created key result "${name}".`,
    };
  },
});
