import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getStoryComments } from "@/modules/story/queries/get-comments";
import { commentStoryAction } from "@/modules/story/actions/comment-story";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { normalizeOptionalString } from "@/lib/ai/tools/normalize-input";
import { MAYA_TOOL_ACTIONS } from "@/lib/ai/tool-actions";

export const commentsTool = tool({
  description:
    "Manage story comments: read comments, add comments, and add threaded replies with user mentions.",
  inputSchema: z.object({
    action: z
      .enum(MAYA_TOOL_ACTIONS.comments)
      .describe("The comment operation to perform"),

    storyId: z.string().describe("Story ID for comment operations"),

    commentId: z
      .string()
      .optional()
      .describe("Comment ID for specific comment operations"),

    parentId: z.string().optional().describe("Parent comment ID for replies"),

    content: z.string().optional().describe("Comment content (HTML supported)"),

    mentions: z
      .array(z.string())
      .optional()
      .describe("Array of user IDs to mention in the comment"),

    includeReplies: z
      .boolean()
      .optional()
      .describe("Include threaded replies in comment responses"),

    limit: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Limit number of comments returned (default: 20, max: 100)"),

    page: z.number().min(1).optional().describe("Page number. Default 1."),

    pageSize: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Comments per page. Default 20, max 100."),
  }),

  execute: async (
    {
      action,
      storyId,
      parentId,
      content,
      mentions = [],
      includeReplies = true,
      limit = 20,
      page,
      pageSize,
    },
    { experimental_context: experimentalContext },
  ) => {
    try {
      const session = await auth();
      const normalizedParentId = normalizeOptionalString(parentId);
      const normalizedContent = normalizeOptionalString(content);

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access comments",
        };
      }

      const workspaceSlug = (experimentalContext as { workspaceSlug: string })
        .workspaceSlug;

      const ctx = { session, workspaceSlug };

      // Get user's workspace and role for permissions
      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;

      switch (action) {
        case "list-comments": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required for listing comments",
            };
          }

          const response = await getStoryComments(
            storyId,
            ctx,
            page ?? 1,
            pageSize ?? limit,
          );
          const comments = response.comments;

          const formattedComments = comments.map((comment) => {
            const commenter = comment.user;
            const replies = includeReplies
              ? comment.subComments.map((reply) => {
                  const replyCommenter = reply.user;
                  return {
                    id: reply.id,
                    content: reply.comment,
                    commenter: {
                      id: replyCommenter.id,
                      name: replyCommenter.fullName,
                      username: replyCommenter.username,
                      avatarUrl: replyCommenter.avatarUrl,
                    },
                    createdAt: reply.createdAt,
                    updatedAt: reply.updatedAt,
                    parentId: reply.parentId,
                  };
                })
              : [];

            return {
              id: comment.id,
              content: comment.comment,
              commenter: {
                id: commenter.id,
                name: commenter.fullName,
                username: commenter.username,
                avatarUrl: commenter.avatarUrl,
              },
              createdAt: comment.createdAt,
              updatedAt: comment.updatedAt,
              parentId: comment.parentId,
              replies,
              replyCount: comment.subComments.length,
            };
          });

          return {
            success: true,
            comments: formattedComments,
            count: formattedComments.length,
            pagination: response.pagination,
            message: `Found ${formattedComments.length} comment${formattedComments.length !== 1 ? "s" : ""} on this story.`,
          };
        }

        case "add-comment": {
          if (!storyId || !normalizedContent) {
            return {
              success: false,
              error: "Story ID and content are required for adding comments",
            };
          }

          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot add comments",
            };
          }

          const result = await commentStoryAction(
            storyId,
            {
              comment: normalizedContent,
              parentId: normalizedParentId ?? null,
              mentions,
            },
            workspaceSlug,
          );

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to add comment",
            };
          }

          const newComment = result.data!;
          const commenter = newComment.user;

          return {
            success: true,
            comment: {
              id: newComment.id,
              content: newComment.comment,
              commenter: {
                id: commenter.id,
                name: commenter.fullName,
                username: commenter.username,
                avatarUrl: commenter.avatarUrl,
              },
              createdAt: newComment.createdAt,
              updatedAt: newComment.updatedAt,
              parentId: newComment.parentId,
            },
            message: normalizedParentId
              ? "Reply added successfully"
              : "Comment added successfully",
          };
        }

        case "reply-to-comment": {
          if (!storyId || !normalizedParentId || !normalizedContent) {
            return {
              success: false,
              error:
                "Story ID, parent comment ID, and content are required for replies",
            };
          }

          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot add replies",
            };
          }

          const result = await commentStoryAction(
            storyId,
            {
              comment: normalizedContent,
              parentId: normalizedParentId,
              mentions,
            },
            workspaceSlug,
          );

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to add reply",
            };
          }

          const newReply = result.data!;
          const commenter = newReply.user;

          return {
            success: true,
            reply: {
              id: newReply.id,
              content: newReply.comment,
              commenter: {
                id: commenter.id,
                name: commenter.fullName,
                username: commenter.username,
                avatarUrl: commenter.avatarUrl,
              },
              createdAt: newReply.createdAt,
              updatedAt: newReply.updatedAt,
              parentId: newReply.parentId,
            },
            message: "Reply added successfully",
          };
        }

        default:
          return {
            success: false,
            error: "Invalid comment action",
          };
      }
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : "An error occurred",
      };
    }
  },
});
