import { z } from "zod";
import { tool } from "ai";
import { headers } from "next/headers";
import { auth } from "@/auth";
import { getStoryComments } from "@/modules/story/queries/get-comments";
import { commentStoryAction } from "@/modules/story/actions/comment-story";
import { getMembers } from "@/lib/queries/members/get-members";

export const commentsTool = tool({
  description:
    "Manage story comments: read, add, edit, delete comments, and handle threaded replies with user mentions. Supports full comment lifecycle management.",
  parameters: z.object({
    action: z
      .enum(["list-comments", "add-comment", "reply-to-comment"])
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
  }),

  execute: async ({
    action,
    storyId,
    parentId,
    content,
    mentions = [],
    includeReplies = true,
    limit = 20,
  }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access comments",
        };
      }

      // Get user's workspace and role for permissions
      const headersList = await headers();
      const subdomain = headersList.get("host")?.split(".")[0] || "";
      const workspace = session.workspaces.find(
        (w) => w.slug.toLowerCase() === subdomain.toLowerCase(),
      );
      const userRole = workspace?.userRole;

      if (!userRole) {
        return {
          success: false,
          error: "Unable to determine user permissions",
        };
      }

      // Get members for mention resolution and commenter info
      const members = await getMembers(session);
      const memberMap = new Map(members.map((m) => [m.id, m]));

      switch (action) {
        case "list-comments": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required for listing comments",
            };
          }

          const comments = await getStoryComments(storyId, session);

          // Filter and limit comments
          const limitedComments = comments.slice(0, limit);

          const formattedComments = limitedComments.map((comment) => {
            const commenter = memberMap.get(comment.userId);
            const replies = includeReplies
              ? comment.subComments.map((reply) => {
                  const replyCommenter = memberMap.get(reply.userId);
                  return {
                    id: reply.id,
                    content: reply.comment,
                    commenter: replyCommenter
                      ? {
                          id: replyCommenter.id,
                          name: replyCommenter.fullName,
                          username: replyCommenter.username,
                          avatarUrl: replyCommenter.avatarUrl,
                        }
                      : null,
                    createdAt: reply.createdAt,
                    updatedAt: reply.updatedAt,
                    parentId: reply.parentId,
                  };
                })
              : [];

            return {
              id: comment.id,
              content: comment.comment,
              commenter: commenter
                ? {
                    id: commenter.id,
                    name: commenter.fullName,
                    username: commenter.username,
                    avatarUrl: commenter.avatarUrl,
                  }
                : null,
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
            totalCount: comments.length,
            message: `Found ${formattedComments.length} comment${formattedComments.length !== 1 ? "s" : ""} on this story.`,
          };
        }

        case "add-comment": {
          if (!storyId || !content) {
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

          const result = await commentStoryAction(storyId, {
            comment: content,
            parentId: parentId ?? null,
            mentions,
          });

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to add comment",
            };
          }

          const newComment = result.data!;
          const commenter = memberMap.get(newComment.userId);

          return {
            success: true,
            comment: {
              id: newComment.id,
              content: newComment.comment,
              commenter: commenter
                ? {
                    id: commenter.id,
                    name: commenter.fullName,
                    username: commenter.username,
                    avatarUrl: commenter.avatarUrl,
                  }
                : null,
              createdAt: newComment.createdAt,
              updatedAt: newComment.updatedAt,
              parentId: newComment.parentId,
            },
            message: parentId
              ? "Reply added successfully"
              : "Comment added successfully",
          };
        }

        case "reply-to-comment": {
          if (!storyId || !parentId || !content) {
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

          const result = await commentStoryAction(storyId, {
            comment: content,
            parentId,
            mentions,
          });

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to add reply",
            };
          }

          const newReply = result.data!;
          const commenter = memberMap.get(newReply.userId);

          return {
            success: true,
            reply: {
              id: newReply.id,
              content: newReply.comment,
              commenter: commenter
                ? {
                    id: commenter.id,
                    name: commenter.fullName,
                    username: commenter.username,
                    avatarUrl: commenter.avatarUrl,
                  }
                : null,
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
