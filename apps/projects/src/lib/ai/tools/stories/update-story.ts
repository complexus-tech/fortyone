import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { updateStoryAction } from "@/modules/story/actions/update-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { normalizeOptionalString } from "@/lib/ai/tools/normalize-input";
import { isEstimateValue } from "@/lib/estimate";
import {
  MAX_TIME_NEEDED_MINUTES,
  normalizeTimeNeededPatch,
} from "@/lib/time-needed";
import { requireToolConfirmation } from "../tool-helpers";

export const updateStory = tool({
  description:
    "Update an existing story. Only admins and members can update stories.",
  inputSchema: z.object({
    storyId: z.string().describe("Story ID to update (required)"),
    confirmed: z
      .boolean()
      .optional()
      .describe(
        "Must be true after the user explicitly confirms the story update.",
      ),
    title: z.string().optional().describe("Updated title"),
    description: z.string().optional().describe("Updated description"),
    descriptionHTML: z.string().optional().describe("Updated description HTML"),
    statusId: z.string().optional().describe("Updated status ID"),
    assigneeId: z.string().optional().describe("Updated assignee ID"),
    priority: z
      .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
      .optional()
      .describe("Updated priority"),
    estimateValue: z
      .number()
      .int()
      .refine(isEstimateValue, {
        message: "Complexity must be 1, 2, 3, 5, or 8.",
      })
      .nullable()
      .optional()
      .describe(
        "Updated relative complexity value for the team's scale. This is not a time duration. Set null to clear complexity.",
      ),
    estimatedDurationMinutes: z
      .number()
      .int()
      .positive()
      .max(MAX_TIME_NEEDED_MINUTES)
      .nullable()
      .optional()
      .describe(
        "Updated total time needed in minutes for scheduling. Set null to clear both the duration and its minimum focus block.",
      ),
    minimumFocusBlockMinutes: z
      .number()
      .int()
      .positive()
      .max(MAX_TIME_NEEDED_MINUTES)
      .nullable()
      .optional()
      .describe(
        "Updated minimum schedulable focus block in minutes. Set null to let Maya automatically fill available calendar time.",
      ),
    autoSchedulingEnabled: z
      .boolean()
      .optional()
      .describe(
        "Enable or pause continuous Maya calendar scheduling for this story.",
      ),
    autoSchedulingLocked: z
      .boolean()
      .optional()
      .describe(
        "Lock or unlock the current Maya calendar blocks. Lock only after blocks have been scheduled.",
      ),
    labelIds: z
      .array(z.string())
      .optional()
      .describe("Replace story labels with these label IDs."),
    sprintId: z.string().optional().describe("Updated sprint ID"),
    objectiveId: z.string().optional().describe("Updated objective ID"),
    startDate: z.string().optional().describe("Updated start date"),
    endDate: z.string().optional().describe("Updated end date"),
  }),

  execute: async (
    {
      storyId,
      confirmed,
      title,
      description,
      descriptionHTML,
      statusId,
      assigneeId,
      priority,
      estimateValue,
      estimatedDurationMinutes,
      minimumFocusBlockMinutes,
      autoSchedulingEnabled,
      autoSchedulingLocked,
      labelIds,
      sprintId,
      objectiveId,
      startDate,
      endDate,
    },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation("update this story");
      }

      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to update stories",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;

      if (userRole === "guest") {
        return {
          success: false,
          error: "Only admins and members can update stories",
        };
      }

      const timeNeededPatch = normalizeTimeNeededPatch(
        estimatedDurationMinutes,
        minimumFocusBlockMinutes,
      );
      const updateData = {
        title: normalizeOptionalString(title),
        description: normalizeOptionalString(description),
        descriptionHTML: normalizeOptionalString(descriptionHTML),
        statusId: normalizeOptionalString(statusId),
        assigneeId: normalizeOptionalString(assigneeId),
        priority,
        estimateValue,
        ...timeNeededPatch,
        autoSchedulingEnabled,
        autoSchedulingLocked,
        labelIds,
        sprintId: normalizeOptionalString(sprintId),
        objectiveId: normalizeOptionalString(objectiveId),
        startDate: normalizeOptionalString(startDate),
        endDate: normalizeOptionalString(endDate),
      };

      const result = await updateStoryAction(
        storyId,
        updateData,
        workspaceSlug,
      );

      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message,
        };
      }

      return {
        success: true,
        message: "Story updated successfully.",
      };
    } catch (error) {
      return {
        success: false,
        error: "Failed to update story",
      };
    }
  },
});
