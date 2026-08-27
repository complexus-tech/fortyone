import { tool } from "ai";
import { auth } from "@/auth";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { isStoryCreationOutcomeUncertainError } from "@/modules/story/actions/story-creation-error";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import {
  normalizeOptionalStoryId,
  normalizeRequiredStoryId,
  normalizeStoryInput,
} from "./normalize-story-input";
import { createSprintEndDateResolver } from "./resolve-sprint-end-date";
import { createStoryStatusResolver } from "./resolve-story-status";
import { getStoryCreationIdempotencyKey } from "./story-creation-idempotency";
import { createStoryInputSchema } from "./story-creation-schema";
import { getStoryCalendarImpact } from "./story-calendar-impact";
import { toStoryToolSummary } from "./story-tool-summary";

export { createStoryInputSchema } from "./story-creation-schema";

export const createStory = tool({
  description:
    "Create one story after its missing planning details have been answered or the user has explicitly chosen to skip them. A future/start/delivery date does not enable calendar scheduling; set autoSchedulingEnabled true only for explicit calendar intent. Guests cannot create stories. Members and admins can create stories for teams they belong to.",
  inputSchema: createStoryInputSchema,
  needsApproval: true,

  execute: async (
    {
      title,
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
      keyResultId,
      parentId,
      startDate,
      endDate,
    },
    { experimental_context: experimentalContext, toolCallId },
  ) => {
    try {
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
      const resolvedSprintId = normalizeOptionalStoryId(sprintId, "sprintId");
      const resolveSprintEndDate = createSprintEndDateResolver(ctx);
      const resolvedEndDate = await resolveSprintEndDate(
        resolvedSprintId,
        endDate,
      );

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
        sprintId: resolvedSprintId,
        objectiveId,
        keyResultId,
        parentId,
        startDate,
        endDate: resolvedEndDate,
      });

      const result = await createStoryAction(
        {
          ...storyData,
          idempotencyKey: getStoryCreationIdempotencyKey({
            context: experimentalContext,
            toolCallId,
          }),
        },
        workspaceSlug,
      );

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

      const calendarImpact = getStoryCalendarImpact(result.data);

      return {
        success: true,
        story: toStoryToolSummary(result.data),
        calendarImpact,
        message: `Story "${title}" created successfully. ${calendarImpact}`,
      };
    } catch (error) {
      if (isStoryCreationOutcomeUncertainError(error)) throw error;

      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to create story",
      };
    }
  },
});
