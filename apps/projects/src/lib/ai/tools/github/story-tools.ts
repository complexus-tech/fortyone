import { tool } from "ai";
import { z } from "zod";
import { deleteStoryGitHubLinkAction } from "@/lib/actions/github/delete-story-github-link";
import { postGitHubCommentAction } from "@/lib/actions/github/post-github-comment";
import { getStoryGitHubComments } from "@/lib/queries/github/get-story-github-comments";
import { getStoryGitHubLinks } from "@/lib/queries/github/get-story-github-links";
import { toStoryLinkSummary } from "./contracts";
import {
  apiErrorMessage,
  getAuthenticatedGitHubContext,
  getStoryDisplayRef,
  requireConfirmation,
  resolveStory,
  toUnexpectedToolError,
} from "./helpers";

const storyReferenceSchema = z.object({
  storyId: z.string().optional().describe("Story ID."),
  storyRef: z
    .string()
    .optional()
    .describe("Story reference such as WEB-123 if story ID is unknown."),
});

const missingStoryReference = () => ({
  success: false,
  error: "Either storyId or storyRef is required.",
});

export const getStoryGitHubLinksTool = tool({
  description:
    "Get GitHub issues, pull requests, branches, or commits linked to a FortyOne story.",
  inputSchema: storyReferenceSchema,
  execute: async ({ storyId, storyRef }, options) => {
    try {
      if (!storyId && !storyRef) {
        return missingStoryReference();
      }

      const ctx = await getAuthenticatedGitHubContext(options);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const story = await resolveStory({ ctx, storyId, storyRef });
      if (!story) {
        return {
          success: false,
          error: "Story not found.",
        };
      }

      const links = await getStoryGitHubLinks(story.id, ctx);
      const displayRef = getStoryDisplayRef(story);

      return {
        success: true,
        kind: "github-story-report",
        title: `${displayRef} GitHub links`,
        story: {
          id: story.id,
          ref: displayRef,
          title: story.title,
        },
        links: links.map(toStoryLinkSummary),
        message: `${displayRef} has ${links.length} GitHub link${links.length === 1 ? "" : "s"}.`,
      };
    } catch (error) {
      return toUnexpectedToolError(error, "Failed to get story GitHub links");
    }
  },
});

export const getStoryGitHubCommentsTool = tool({
  description:
    "Get comments from the GitHub issue or pull request linked to a FortyOne story.",
  inputSchema: storyReferenceSchema,
  execute: async ({ storyId, storyRef }, options) => {
    try {
      if (!storyId && !storyRef) {
        return missingStoryReference();
      }

      const ctx = await getAuthenticatedGitHubContext(options);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const story = await resolveStory({ ctx, storyId, storyRef });
      if (!story) {
        return {
          success: false,
          error: "Story not found.",
        };
      }

      const comments = await getStoryGitHubComments(story.id, ctx);
      const displayRef = getStoryDisplayRef(story);

      return {
        success: true,
        story: {
          id: story.id,
          ref: displayRef,
          title: story.title,
        },
        comments: comments.map((comment) => ({
          id: comment.id,
          body: comment.body,
          userLogin: comment.userLogin,
          userAvatar: comment.userAvatar,
          createdAt: comment.createdAt,
          updatedAt: comment.updatedAt,
          htmlUrl: comment.htmlUrl,
        })),
        message: `Found ${comments.length} GitHub comment${comments.length === 1 ? "" : "s"} for ${displayRef}.`,
      };
    } catch (error) {
      return toUnexpectedToolError(
        error,
        "Failed to get story GitHub comments",
      );
    }
  },
});

export const postStoryGitHubCommentTool = tool({
  description:
    "Post a comment to the GitHub issue or pull request linked to a FortyOne story.",
  inputSchema: storyReferenceSchema.extend({
    body: z.string().min(1).describe("Comment body to post to GitHub."),
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms posting."),
  }),
  execute: async ({ storyId, storyRef, body, confirmed }, options) => {
    try {
      if (!confirmed) {
        return requireConfirmation("post this comment to GitHub");
      }

      if (!storyId && !storyRef) {
        return missingStoryReference();
      }

      const ctx = await getAuthenticatedGitHubContext(options);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const story = await resolveStory({ ctx, storyId, storyRef });
      if (!story) {
        return {
          success: false,
          error: "Story not found.",
        };
      }

      const result = await postGitHubCommentAction(
        story.id,
        { body },
        ctx.workspaceSlug,
      );
      const displayRef = getStoryDisplayRef(story);

      if (result.error) {
        return {
          success: false,
          error: apiErrorMessage(result, "Failed to post GitHub comment"),
        };
      }

      return {
        success: true,
        message: `Posted the comment to GitHub for ${displayRef}.`,
      };
    } catch (error) {
      return toUnexpectedToolError(error, "Failed to post GitHub comment");
    }
  },
});

export const deleteStoryGitHubLinkTool = tool({
  description:
    "Remove a GitHub link from a FortyOne story. Use getStoryGitHubLinksTool first to find the link.",
  inputSchema: storyReferenceSchema.extend({
    confirmed: z
      .boolean()
      .optional()
      .describe("Must be true after the user explicitly confirms unlinking."),
    linkId: z.string().describe("Story GitHub link ID."),
  }),
  execute: async ({ storyId, storyRef, linkId, confirmed }, options) => {
    try {
      if (!confirmed) {
        return requireConfirmation("remove this GitHub link from the story");
      }

      if (!storyId && !storyRef) {
        return missingStoryReference();
      }

      const ctx = await getAuthenticatedGitHubContext(options);
      if ("error" in ctx) return { success: false, error: ctx.error };

      const story = await resolveStory({ ctx, storyId, storyRef });
      if (!story) {
        return {
          success: false,
          error: "Story not found.",
        };
      }

      const result = await deleteStoryGitHubLinkAction(
        story.id,
        linkId,
        ctx.workspaceSlug,
      );
      const displayRef = getStoryDisplayRef(story);

      if (result.error) {
        return {
          success: false,
          error: apiErrorMessage(result, "Failed to remove GitHub link"),
        };
      }

      return {
        success: true,
        message: `Removed the GitHub link from ${displayRef}.`,
      };
    } catch (error) {
      return toUnexpectedToolError(error, "Failed to remove story GitHub link");
    }
  },
});
