import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { updateKeyResult } from "@/modules/objectives/actions/update-key-result";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const updateKeyResultTool = tool({
  description:
    "Update existing key results. Can update name, measurement type, and values.",
  inputSchema: z.object({
    keyResultId: z.string().describe("Key result ID to update"),
    name: z.string().optional().describe("Updated key result name"),
    startValue: z.number().optional().describe("Updated starting value"),
    currentValue: z
      .number()
      .optional()
      .describe("Updated current progress value"),
    targetValue: z.number().optional().describe("Updated target value"),
    startDate: z
      .string()
      .optional()
      .describe("Updated start date (ISO date string e.g 2005-06-13)"),
    endDate: z
      .string()
      .optional()
      .describe("Updated end date (ISO date string e.g 2005-06-13)"),
    lead: z.string().optional().describe("Updated lead user ID (optional)"),
    contributors: z
      .array(z.string())
      .optional()
      .describe("Updated contributors user IDs (optional)"),
  }),

  execute: async ({
    keyResultId,
    name,
    startValue,
    currentValue,
    targetValue,
    startDate,
    endDate,
    lead,
    contributors,
  }) => {
    const session = await auth();

    if (!session) {


      return {
        success: false,
        error: "Authentication required",
      };
    }

    // Get user's workspace and role for permissions
    const workspace = await getWorkspace(session);
    const userRole = workspace.userRole;

    if (userRole === "guest") {
      return {
        success: false,
        error: "Guests cannot update key results",
      };
    }

    const result = await updateKeyResult(keyResultId, {
      name,
      startValue,
      currentValue,
      targetValue,
      startDate,
      endDate,
      lead,
      contributors,
    });

    if (result.error) {
      return {
        success: false,
        error: result.error.message || "Failed to update key result",
      };
    }

    return {
      success: true,
      message: "Successfully updated key result.",
    };
  },
});
