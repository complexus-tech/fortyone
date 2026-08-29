import { tool } from "ai";
import { auth } from "@/auth";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { isStoryCreationOutcomeUncertainError } from "@/modules/story/actions/story-creation-error";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import type { DetailedStory } from "@/modules/story/types";
import {
  normalizeOptionalStoryId,
  normalizeRequiredStoryId,
  normalizeStoryInput,
} from "./normalize-story-input";
import { createSprintEndDateResolver } from "./resolve-sprint-end-date";
import { createStoryStatusResolver } from "./resolve-story-status";
import { getStoryCreationIdempotencyKey } from "./story-creation-idempotency";
import {
  applyBulkStorySharedValues,
  bulkCreateStoriesInputSchema,
} from "./story-creation-schema";
import { getBulkStoryCalendarImpact } from "./story-calendar-impact";
import { toStoryToolSummary } from "./story-tool-summary";

const CREATION_BATCH_SIZE = 10;

type FailedStory = {
  kind: "failed";
  title: string;
  error: string;
};

type PreparedStory = {
  index: number;
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

export { bulkCreateStoriesInputSchema } from "./story-creation-schema";

export const bulkCreateStories = tool({
  description:
    "Bulk create up to 50 stories. Default every story to no time estimate and calendar scheduling off; do not ask for or apply one batch-wide duration. Put team, assignee, status, or other genuinely common metadata in sharedValues. Include shared planning values only when the user explicitly says they apply to every story; otherwise use only supplied per-story planning values. Execution pauses for approval, processes in batches of 10, and reports every result.",
  inputSchema: bulkCreateStoriesInputSchema,
  needsApproval: true,

  execute: async (
    { sharedValues, storiesData },
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
        storiesData.map(async (storyInput, index) => {
          const storyData = applyBulkStorySharedValues(
            sharedValues,
            storyInput,
          );
          const title = storyData.title.trim() || `Story ${index + 1}`;

          try {
            const teamId = normalizeRequiredStoryId(
              storyData.teamId ?? "",
              "teamId",
            );
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
              index,
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
          batch.map(async ({ index, story }) => {
            const result = await createStoryAction(
              {
                ...story,
                idempotencyKey: getStoryCreationIdempotencyKey({
                  context: experimentalContext,
                  index,
                  toolCallId,
                }),
              },
              workspaceSlug,
            );

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

      const calendarImpact = getBulkStoryCalendarImpact(createdStories);

      if (errorCount > 0) {
        return {
          success: false,
          createdCount: successCount,
          errorCount,
          stories: createdStories.map(toStoryToolSummary),
          failedStories,
          calendarImpact,
          error: failedStories
            .map((failure) => `${failure.title}: ${failure.error}`)
            .join("; "),
          message: `Created ${successCount} stories. ${errorCount} stories failed to create. ${calendarImpact}`,
        };
      }

      return {
        success: true,
        createdCount: successCount,
        errorCount,
        stories: createdStories.map(toStoryToolSummary),
        calendarImpact,
        message: `Successfully created ${successCount} stories. ${calendarImpact}`,
      };
    } catch (error) {
      if (isStoryCreationOutcomeUncertainError(error)) throw error;

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
