import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getKeyResultActivities } from "@/modules/objectives/queries/get-key-result-activities";

export const getKeyResultActivitiesTool = tool({
  description:
    "View key result activity timeline and changes: track who made what changes, when they happened, and provide detailed history of key result modifications.",
  inputSchema: z.object({
    keyResultId: z.string().describe("Key Result ID for activity operations"),
    page: z.number().min(1).optional().describe("Page number for pagination"),
    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Number of activities per page"),
  }),

  execute: async ({ keyResultId, page = 1, pageSize = 20 }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access key result activities",
        };
      }

      const response = await getKeyResultActivities(
        keyResultId,
        session,
        page,
        pageSize,
      );
      const activities = response.activities;

      return {
        success: true,
        activities,
        count: activities.length,
        pagination: response.pagination,
        message: `Found ${activities.length} activity${activities.length !== 1 ? "s" : ""} for this key result.`,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "An error occurred",
      };
    }
  },
});
