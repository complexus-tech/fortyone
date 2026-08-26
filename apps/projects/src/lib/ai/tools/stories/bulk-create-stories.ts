import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { isEstimateValue } from "@/lib/estimate";
import { MAX_TIME_NEEDED_MINUTES } from "@/lib/time-needed";
import type { DetailedStory } from "@/modules/story/types";
import {
  normalizeOptionalStoryId,
  normalizeRequiredStoryId,
  normalizeStoryInput,
} from "./normalize-story-input";
import { createSprintEndDateResolver } from "./resolve-sprint-end-date";
import { createStoryStatusResolver } from "./resolve-story-status";
import { toStoryToolSummary } from "./story-tool-summary";

const MAX_STORIES_PER_REQUEST = 50;
const CREATION_BATCH_SIZE = 10;

type FailedStory = {
  kind: "failed";
  title: string;
  error: string;
};

type PreparedStory = {
  kind: "prepared";
  story: ReturnType<typeof normalizeStoryInput>;
};

type PreparationResult = FailedStory | PreparedStory;

type CreationResult =
  | FailedStory
  | {
      kind: "created";
      story: DetailedStory;
    };

const getErrorMessage = (error: unknown, fallback: string) =>
  error instanceof Error ? error.message : fallback;

const inBatchesOf = <Item>(items: Item[], size: number) => {
  const batches: Item[][] = [];
  for (let index = 0; index < items.length; index += size) {
    batches.push(items.slice(index, index + size));
  }
  return batches;
};

export const bulkCreateStoriesInputSchema = z.object({
  storiesData: z
    .array(
      z.object({
        title: z.string().describe("Story title (required)"),
        description: z
          .string()
          .nullable()
          .optional()
          .describe("Story description"),
        descriptionHTML: z
          .string()
          .nullable()
          .optional()
          .describe("Story description HTML"),
        teamId: z.string().describe("Team ID where story belongs (required)"),
        statusId: z
          .string()
          .nullable()
          .optional()
          .describe(
            "Initial status ID. Omit it to use the team's default status.",
          ),
        assigneeId: z
          .string()
          .nullable()
          .optional()
          .describe("Assignee user ID"),
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
            "Optional smallest schedulable focus block in minutes. It cannot exceed estimatedDurationMinutes.",
          ),
        autoSchedulingEnabled: z
          .boolean()
          .optional()
          .describe(
            "Whether Maya should continuously schedule this story. Defaults to false for human assignees; set true when requested or when assigning to Maya.",
          ),
        labelIds: z
          .array(z.string())
          .nullable()
          .optional()
          .describe("Label IDs to attach to the story."),
        sprintId: z
          .string()
          .nullable()
          .optional()
          .describe("Sprint ID to assign story"),
        objectiveId: z
          .string()
          .nullable()
          .optional()
          .describe("Objective ID to assign story"),
        keyResultId: z
          .string()
          .nullable()
          .optional()
          .describe("Key result ID to assign story"),
        parentId: z
          .string()
          .nullable()
          .optional()
          .describe("Parent story ID for sub-stories"),
        startDate: z
          .string()
          .nullable()
          .optional()
          .describe("Story start date (ISO  date string e.g 2005-06-13)"),
        endDate: z
          .string()
          .nullable()
          .optional()
          .describe("Story end date (ISO  date string e.g 2005-06-13)"),
      }),
    )
    .min(1, "Provide at least one story to create.")
    .max(
      MAX_STORIES_PER_REQUEST,
      `Create at most ${MAX_STORIES_PER_REQUEST} stories in one request.`,
    )
    .describe("Array of story data for bulk creation (required)"),
});

export const bulkCreateStories = tool({
  description:
    "Bulk create multiple stories at once. Only admins and members can perform bulk operations. Prepare up to 50 stories in one request; execution pauses for user approval, then processes them safely in batches of 10 and reports every success or failure.",
  inputSchema: bulkCreateStoriesInputSchema,
  needsApproval: true,

  execute: async (
    { storiesData },
    { experimental_context: experimentalContext },
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

      // Only admins can perform bulk operations
      if (userRole === "guest") {
        return {
          success: false,
          error: "Only admins and members can perform bulk story operations",
        };
      }

      const resolveStatusId = createStoryStatusResolver(ctx);
      const resolveSprintEndDate = createSprintEndDateResolver(ctx);
      const preparedResults: PreparationResult[] = await Promise.all(
        storiesData.map(async (storyData, index) => {
          const title = storyData.title.trim() || `Story ${index + 1}`;

          try {
            const teamId = normalizeRequiredStoryId(storyData.teamId, "teamId");
            const statusId = await resolveStatusId(teamId, storyData.statusId);
            const sprintId = normalizeOptionalStoryId(
              storyData.sprintId,
              "sprintId",
            );
            const endDate = await resolveSprintEndDate(
              sprintId,
              storyData.endDate,
            );

            return {
              kind: "prepared",
              story: normalizeStoryInput({
                ...storyData,
                teamId,
                statusId,
                sprintId,
                endDate,
              }),
            };
          } catch (error) {
            return {
              error: getErrorMessage(error, "Failed to validate story data."),
              kind: "failed",
              title,
            };
          }
        }),
      );

      const failedStories: FailedStory[] = preparedResults.flatMap((result) =>
        result.kind === "failed"
          ? [{ error: result.error, kind: "failed", title: result.title }]
          : [],
      );
      const preparedStories = preparedResults.flatMap((result) =>
        result.kind === "prepared" ? [result] : [],
      );
      const createdStories: DetailedStory[] = [];

      for (const batch of inBatchesOf(preparedStories, CREATION_BATCH_SIZE)) {
        // eslint-disable-next-line no-await-in-loop -- Limits concurrent creation mutations to a safe batch.
        const results: CreationResult[] = await Promise.all(
          batch.map(async ({ story }) => {
            const result = await createStoryAction(story, workspaceSlug);

            if (result.error?.message) {
              return {
                error: result.error.message,
                kind: "failed",
                title: story.title,
              };
            }

            if (!result.data?.id) {
              return {
                error: "Story creation did not return a created story.",
                kind: "failed",
                title: story.title,
              };
            }

            return { kind: "created", story: result.data };
          }),
        );

        for (const result of results) {
          if (result.kind === "failed") {
            failedStories.push({
              error: result.error,
              kind: "failed",
              title: result.title,
            });
          } else {
            createdStories.push(result.story);
          }
        }
      }

      const successCount = createdStories.length;
      const errorCount = failedStories.length;

      if (errorCount > 0) {
        return {
          success: false,
          createdCount: successCount,
          errorCount,
          stories: createdStories.map(toStoryToolSummary),
          failedStories,
          error: failedStories
            .map((failure) => `${failure.title}: ${failure.error}`)
            .join("; "),
          message: `Created ${successCount} stories. ${errorCount} stories failed to create.`,
        };
      }

      return {
        success: true,
        createdCount: successCount,
        errorCount,
        stories: createdStories.map(toStoryToolSummary),
        message: `Successfully created ${successCount} stories.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to bulk create stories",
      };
    }
  },
});
