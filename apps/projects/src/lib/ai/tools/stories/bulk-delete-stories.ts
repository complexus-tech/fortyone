import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { isStoryDeletionOutcomeUncertainError } from "@/modules/story/actions/story-deletion-error";
import { bulkDeleteAction } from "@/modules/stories/actions/bulk-delete-stories";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { requireToolConfirmation } from "../tool-helpers";

export const bulkDeleteStories = tool({
  description:
    "Bulk delete multiple stories at once. Resolve and provide every exact story ID and matching human-readable title so the approval screen can show verifiable targets. Only admins and members can perform bulk operations.",
  inputSchema: z
    .object({
      storyIds: z
        .array(z.string().uuid("Each story ID must be a valid UUID."))
        .min(1, "Provide at least one story to delete.")
        .max(50, "Delete at most 50 stories in one request.")
        .describe("Exact story IDs to delete, ordered to match storyTitles."),
      storyTitles: z
        .array(z.string().trim().min(1, "Each story title is required."))
        .min(1, "Provide every story title for approval.")
        .max(50, "Delete at most 50 stories in one request.")
        .describe(
          "Human-readable title for each storyIds entry, in the same order. These titles are shown to the user before deletion.",
        ),
      confirmed: z
        .boolean()
        .optional()
        .describe(
          "Must be true after the user explicitly confirms the bulk deletion.",
        ),
    })
    .superRefine(({ storyIds, storyTitles }, ctx) => {
      if (storyIds.length === storyTitles.length) return;

      ctx.addIssue({
        code: "custom",
        message: "Provide one matching title for every story ID.",
        path: ["storyTitles"],
      });
    }),

  execute: async (
    { storyIds, confirmed },
    { experimental_context: experimentalContext },
  ) => {
    try {
      if (!confirmed) {
        return requireToolConfirmation("delete these stories");
      }

      const session = await auth();
      if (!session) {
        return {
          success: false,
          error: "Authentication required to bulk delete stories",
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
          error: "Only admins and members can perform bulk story operations",
        };
      }

      const result = await bulkDeleteAction({ storyIds }, workspaceSlug);

      if (result.error?.message) {
        return {
          success: false,
          error: result.error.message || "Failed to bulk delete stories",
        };
      }
      const deletedCount = result.data?.deletedCount ?? 0;
      const deletedStoryIds = result.data?.storyIds ?? [];
      const deletedStoryIdSet = new Set(deletedStoryIds);
      const missingStoryIds = storyIds.filter(
        (storyId) => !deletedStoryIdSet.has(storyId),
      );
      const completed =
        deletedCount === storyIds.length && missingStoryIds.length === 0;

      return {
        success: completed,
        deletedCount,
        requestedCount: storyIds.length,
        storyIds: deletedStoryIds,
        missingStoryIds,
        message: completed
          ? `Successfully deleted ${deletedCount} stories.`
          : `Deleted ${deletedCount} of ${storyIds.length} stories. ${missingStoryIds.length} could not be deleted because they were missing, already deleted, or unavailable in this workspace.`,
      };
    } catch (error) {
      if (isStoryDeletionOutcomeUncertainError(error)) throw error;

      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to bulk delete stories",
      };
    }
  },
});
