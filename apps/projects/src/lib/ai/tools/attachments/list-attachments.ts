import { z } from "zod";
import { tool } from "ai";
import { auth } from "@/auth";
import { getStoryAttachments } from "@/modules/story/queries/get-attachments";

export const listAttachments = tool({
  description:
    "List all attachments for a specific story. Returns file information, uploader details, and summary statistics.",
  inputSchema: z.object({
    storyId: z.string().describe("Story ID to list attachments for (required)"),
  }),

  execute: async ({ storyId }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access attachments",
        };
      }

      const attachments = await getStoryAttachments(storyId, session);

      const formatFileSize = (bytes: number) => {
        if (bytes < 1024) return `${bytes} B`;
        if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`;
        return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
      };

      const formattedAttachments = attachments.map((attachment) => {
        return {
          id: attachment.id,
          filename: attachment.filename,
          size: attachment.size,
          formattedSize: formatFileSize(attachment.size),
          mimeType: attachment.mimeType,
          url: attachment.url,
          createdAt: attachment.createdAt,
          uploadedBy: attachment.uploadedBy,
        };
      });

      return {
        success: true,
        attachments: formattedAttachments,
        count: formattedAttachments.length,
        totalCount: formattedAttachments.length,
        message: `Found ${formattedAttachments.length} attachment${formattedAttachments.length !== 1 ? "s" : ""} on this story.`,
      };
    } catch (error) {
      return {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to list attachments",
      };
    }
  },
});
