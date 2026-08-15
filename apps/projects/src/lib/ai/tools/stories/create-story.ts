import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { isEstimateValue } from "@/lib/estimate";
import { MAX_TIME_NEEDED_MINUTES } from "@/lib/time-needed";
import { requireToolConfirmation } from "../tool-helpers";
import {
  normalizeRequiredStoryId,
  normalizeStoryInput,
} from "./normalize-story-input";
import { createStoryStatusResolver } from "./resolve-story-status";

export const createStory = tool({
  description:
    "Create a new story. Guests cannot create stories. Members and admins can create stories for teams they belong to.",
  inputSchema: z.object({
    title: z.string().describe("Story title (required)"),
    confirmed: z
      .boolean()
      .optional()
      .describe(
        "Must be true after the user explicitly confirms creating the story.",
      ),
    description: z.string().nullable().optional().describe("Story description"),
    descriptionHTML: z
      .string()
      .nullable()
      .optional()
      .describe(
        "Story description HTML (Always provided and properly formatted if description is provided)",
      ),
    teamId: z
      .string()
      .describe("Team ID where story belongs (required) (UUID)"),
    statusId: z
      .string()
      .nullable()
      .optional()
      .describe(
        "Initial status ID (UUID). Resolve it with the statuses tool when the user specifies a status; otherwise omit it to use the team's default status.",
      ),
    assigneeId: z
      .string()
      .nullable()
      .optional()
      .describe("Assignee user ID (UUID)"),
    priority: z
      .enum(["No Priority", "Low", "Medium", "High", "Urgent"])
      .default("No Priority")
      .describe("Story priority (required)"),
    estimateValue: z
      .number()
      .int()
      .refine((value) => value === 0 || isEstimateValue(value), {
        message: "Complexity must be 1, 2, 3, 5, or 8.",
      })
      .nullable()
      .optional()
      .describe(
        "Relative complexity value using the team's scale. Use 1, 2, 3, 5, or 8. This is not a time duration; use 0, null, or omit when unset.",
      ),
    estimatedDurationMinutes: z
      .number()
      .int()
      .positive()
      .max(MAX_TIME_NEEDED_MINUTES)
      .nullable()
      .optional()
      .describe(
        "Total time needed in minutes for calendar scheduling. Omit or set null when unknown.",
      ),
    minimumFocusBlockMinutes: z
      .number()
      .int()
      .positive()
      .max(MAX_TIME_NEEDED_MINUTES)
      .nullable()
      .optional()
      .describe(
        "Optional smallest schedulable focus block in minutes. It cannot exceed estimatedDurationMinutes; omit for the automatic default.",
      ),
    autoSchedulingEnabled: z
      .boolean()
      .optional()
      .describe(
        "Whether Maya should continuously place this story on the assignee's calendar. Defaults to false for human-assigned stories; set true when the user requests auto-scheduling or assigns the story to Maya.",
      ),
    labelIds: z
      .array(z.string())
      .optional()
      .describe("Label IDs to attach to the story."),
    sprintId: z
      .string()
      .nullable()
      .optional()
      .describe("Sprint ID to assign story (UUID)"),
    objectiveId: z
      .string()
      .nullable()
      .optional()
      .describe("Objective ID to assign story (UUID)"),
    parentId: z
      .string()
      .nullable()
      .optional()
      .describe("Parent story ID for sub-stories (UUID)"),
    startDate: z
      .string()
      .nullable()
      .optional()
      .describe("Story start date (ISO date string e.g 2005-06-13)"),
    endDate: z
      .string()
      .nullable()
      .optional()
      .describe("Story end date (ISO date string e.g 2005-06-13)"),
  }),

  execute: async (
    {
      title,
      confirmed,
      description,
      descriptionHTML,
      teamId,
      statusId,
      assigneeId,
      priority,
      estimateValue,
      estimatedDurationMinutes,
      minimumFocusBlockMinutes,
      autoSchedulingEnabled,
      labelIds,
      sprintId,
      objectiveId,
      parentId,
      startDate,
      endDate,
    },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation("create this story");
      }

      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to create stories",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;

      // Check permissions for guests
      if (userRole === "guest") {
        return {
          success: false,
          error: "Guests can only create stories for teams they belong to",
        };
      }

      const resolvedTeamId = normalizeRequiredStoryId(teamId, "teamId");
      const resolveStatusId = createStoryStatusResolver(ctx);
      const resolvedStatusId = await resolveStatusId(resolvedTeamId, statusId);

      const storyData = normalizeStoryInput({
        title,
        description,
        descriptionHTML,
        teamId: resolvedTeamId,
        statusId: resolvedStatusId,
        assigneeId,
        priority,
        estimateValue,
        estimatedDurationMinutes,
        minimumFocusBlockMinutes,
        autoSchedulingEnabled,
        labelIds,
        sprintId,
        objectiveId,
        parentId,
        startDate,
        endDate,
      });

      const result = await createStoryAction(storyData, workspaceSlug);

      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message || "Failed to create story",
        };
      }

      if (!result.data?.id) {
        return {
          success: false,
          error: "Story creation did not return a created story.",
        };
      }

      return {
        success: true,
        story: result.data,
        message: `Story "${title}" created successfully.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to create story",
      };
    }
  },
});
