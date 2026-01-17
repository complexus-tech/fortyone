import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { deleteStoryAttachmentAction } from "@/modules/story/actions/delete-attachment";
import { getStoryAttachments } from "@/modules/story/queries/get-attachments";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";

export const deleteAttachment = tool({
  description:
    "Delete a specific attachment from a story. Only admins or the attachment uploader can delete attachments.",
  inputSchema: z.object({
    storyId: z
      .string()
      .describe("Story ID that contains the attachment (required)"),
    attachmentId: z.string().describe("Attachment ID to delete (required)"),
  }),

  execute: async ({ storyId, attachmentId }, { experimental_context }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to delete attachments",
        };
      }

      const workspaceSlug = (experimental_context as { workspaceSlug: string }).workspaceSlug;
      

      const ctx = { session, workspaceSlug };

      const workspace = await getWorkspace(ctx);
      const userRole = workspace.userRole;
      const userId = session.user!.id;

      // Check if user owns the attachment or is admin
      const attachments = await getStoryAttachments(storyId, ctx);
      const attachment = attachments.find((a) => a.id === attachmentId);

      if (!attachment) {
        return {
          success: false,
          error: "Attachment not found",
        };
      }

      if (attachment.uploadedBy !== userId && userRole !== "admin") {
        return {
          success: false,
          error: "You can only delete your own attachments",
        };
      }

      const result = await deleteStoryAttachmentAction(storyId, attachmentId, workspaceSlug);

      if (result.error) {
        return {
          success: false,
          error: result.error.message || "Failed to delete attachment",
        };
      }

      return {
        success: true,
        message: `Attachment "${attachment.filename}" deleted successfully`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error
            ? error.message
            : "Failed to delete attachment",
      };
    }
  },
});
