import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { bulkUpdateAction } from "@/modules/stories/actions/bulk-update-stories";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { normalizeOptionalString } from "@/lib/ai/tools/normalize-input";
import { isEstimateValue } from "@/lib/estimate";
import {
  MAX_TIME_NEEDED_MINUTES,
  normalizeTimeNeededPatch,
} from "@/lib/time-needed";
import { requireToolConfirmation } from "../tool-helpers";

export const bulkUpdateStories = tool({
  description:
    "Bulk update multiple stories at once. Only admins and members can perform bulk operations.",
  inputSchema: z.object({
    storyIds: z
      .array(z.string().uuid("Each story ID must be a valid UUID."))
      .min(1, "Provide at least one story to update.")
      .max(50, "Update at most 50 stories in one request.")
      .describe("Array of story IDs to update (required)"),
    confirmed: z
      .boolean()
      .optional()
      .describe(
        "Must be true after the user explicitly confirms the bulk update.",
      ),
    updateData: z
      .object({
        statusId: z
          .string()
          .optional()
          .describe("Updated status ID for all stories"),
        assigneeId: z
          .string()
          .optional()
          .describe("Updated assignee ID for all stories"),
        priority: z
          .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
          .optional()
          .describe("Updated priority for all stories"),
        sprintId: z
          .string()
          .optional()
          .describe("Updated sprint ID for all stories"),
        objectiveId: z
          .string()
          .optional()
          .describe("Updated objective ID for all stories"),
        startDate: z
          .string()
          .optional()
          .describe(
            "Updated start date for all stories (ISO date string e.g 2005-06-13)",
          ),
        endDate: z
          .string()
          .optional()
          .describe(
            "Updated end date for all stories (ISO date string e.g 2005-06-13)",
          ),
        estimateValue: z
          .number()
          .int()
          .refine(isEstimateValue, {
            message: "Complexity must be 1, 2, 3, 5, or 8.",
          })
          .nullable()
          .optional()
          .describe(
            "Updated relative complexity value for all stories. This is not a time duration. Set null to clear complexity.",
          ),
        estimatedDurationMinutes: z
          .number()
          .int()
          .positive()
          .max(MAX_TIME_NEEDED_MINUTES)
          .nullable()
          .optional()
          .describe(
            "Updated total time needed in minutes for all selected stories. Set null to clear both the duration and minimum focus block.",
          ),
        minimumFocusBlockMinutes: z
          .number()
          .int()
          .positive()
          .max(MAX_TIME_NEEDED_MINUTES)
          .nullable()
          .optional()
          .describe(
            "Updated minimum focus block in minutes. Set null to let Maya automatically fill available calendar time.",
          ),
        autoSchedulingEnabled: z
          .boolean()
          .optional()
          .describe(
            "Enable or pause continuous Maya calendar scheduling for every selected story.",
          ),
      })
      .refine(
        (updateData) => Object.keys(updateData).length > 0,
        "Provide at least one field to update.",
      )
      .describe("Update data to apply to all stories (required)"),
  }),

  execute: async (
    { storyIds, confirmed, updateData },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation("bulk update these stories");
      }

      const session = await auth();
      if (!session) {
        return {
          success: false,
          error: "Authentication required to bulk update stories",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;

      // Only admins can perform bulk operations
      if (userRole === "guest") {
        return {
          success: false,
          error: "Only admins and members can perform bulk story operations",
        };
      }

      const timeNeededPatch = normalizeTimeNeededPatch(
        updateData.estimatedDurationMinutes,
        updateData.minimumFocusBlockMinutes,
      );
      const normalizedUpdateData = {
        statusId: normalizeOptionalString(updateData.statusId),
        assigneeId: normalizeOptionalString(updateData.assigneeId),
        priority: updateData.priority,
        sprintId: normalizeOptionalString(updateData.sprintId),
        objectiveId: normalizeOptionalString(updateData.objectiveId),
        startDate: normalizeOptionalString(updateData.startDate),
        endDate: normalizeOptionalString(updateData.endDate),
        estimateValue: updateData.estimateValue,
        ...timeNeededPatch,
        autoSchedulingEnabled: updateData.autoSchedulingEnabled,
      };

      const result = await bulkUpdateAction(
        {
          storyIds,
          updates: normalizedUpdateData,
        },
        workspaceSlug,
      );

      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message || "Failed to bulk update stories",
        };
      }

      if (!result.data) {
        return {
          success: false,
          error: "The bulk update completed without an itemized result",
        };
      }

      const completedSuccessfully = result.data.failedCount === 0;
      const message = completedSuccessfully
        ? `Updated ${result.data.succeededCount} of ${result.data.totalCount} stories.`
        : `Updated ${result.data.succeededCount} of ${result.data.totalCount} stories; ${result.data.failedCount} failed.`;

      return {
        success: completedSuccessfully,
        partial: result.data.partial,
        result: result.data,
        message,
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to bulk update stories",
      };
    }
  },
});
