import { z } from "zod";
import { tool } from "ai";
import { headers } from "next/headers";
import { auth } from "@/auth";
import { getStoryAttachments } from "@/modules/story/queries/get-attachments";
import { addAttachmentAction } from "@/modules/story/actions/add-attachment";
import { deleteStoryAttachmentAction } from "@/modules/story/actions/delete-attachment";
import { getMembers } from "@/lib/queries/members/get-members";

export const attachmentsTool = tool({
  description:
    "Manage story attachments: list, upload, delete files associated with stories. Supports images and PDFs.",
  parameters: z.object({
    action: z
      .enum(["list-attachments", "upload-attachment", "delete-attachment"])
      .describe("The attachment operation to perform"),

    storyId: z.string().describe("Story ID for attachment operations"),

    attachmentId: z
      .string()
      .optional()
      .describe("Attachment ID for specific attachment operations"),

    fileData: z
      .object({
        name: z.string().describe("File name"),
        size: z.number().describe("File size in bytes"),
        type: z.string().describe("File MIME type"),
        content: z.string().describe("File content as base64 string"),
      })
      .optional()
      .describe("File data for upload operations"),

    includeMetadata: z
      .boolean()
      .optional()
      .describe("Include detailed metadata in responses"),

    limit: z
      .number()
      .min(1)
      .max(100)
      .optional()
      .describe("Limit number of attachments returned (default: 20, max: 100)"),
  }),

  execute: async ({
    action,
    storyId,
    attachmentId,
    fileData,
    includeMetadata = false,
    limit = 20,
  }) => {
    try {
      const session = await auth();

      if (!session) {
        return {
          success: false,
          error: "Authentication required to access attachments",
        };
      }

      // Get user's workspace and role for permissions
      const headersList = await headers();
      const subdomain = headersList.get("host")?.split(".")[0] || "";
      const workspace = session.workspaces.find(
        (w) => w.slug.toLowerCase() === subdomain.toLowerCase(),
      );
      const userRole = workspace?.userRole;
      const userId = session.user!.id;

      if (!userRole) {
        return {
          success: false,
          error: "Unable to determine user permissions",
        };
      }

      // Get members for uploader resolution and metadata
      const members = await getMembers(session);
      const memberMap = new Map(members.map((m) => [m.id, m]));

      const formatFileSize = (bytes: number) => {
        if (bytes < 1024) return `${bytes} B`;
        if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`;
        return `${(bytes / (1024 * 1024)).toFixed(2)} MB`;
      };

      const getFileType = (mimeType: string) => {
        if (mimeType.startsWith("image/")) return "image";
        return "pdf";
      };

      switch (action) {
        case "list-attachments": {
          if (!storyId) {
            return {
              success: false,
              error: "Story ID is required for listing attachments",
            };
          }

          const attachments = await getStoryAttachments(storyId, session);

          // Filter and limit attachments
          const limitedAttachments = attachments.slice(0, limit);

          const formattedAttachments = limitedAttachments.map((attachment) => {
            const uploader = memberMap.get(attachment.uploadedBy);
            const baseAttachment = {
              id: attachment.id,
              filename: attachment.filename,
              size: attachment.size,
              formattedSize: formatFileSize(attachment.size),
              mimeType: attachment.mimeType,
              fileType: getFileType(attachment.mimeType),
              url: attachment.url,
              createdAt: attachment.createdAt,
              uploader: uploader
                ? {
                    id: uploader.id,
                    name: uploader.fullName,
                    username: uploader.username,
                    avatarUrl: uploader.avatarUrl,
                  }
                : null,
            };

            if (includeMetadata) {
              return {
                ...baseAttachment,
                isImage: attachment.mimeType.startsWith("image/"),
                isPdf: attachment.mimeType.includes("pdf"),
              };
            }

            return baseAttachment;
          });

          // Group by file type
          const images = formattedAttachments.filter(
            (a) => a.fileType === "image",
          );
          const pdfs = formattedAttachments.filter((a) => a.fileType === "pdf");

          return {
            success: true,
            attachments: formattedAttachments,
            count: formattedAttachments.length,
            totalCount: attachments.length,
            summary: {
              images: images.length,
              pdfs: pdfs.length,
              totalSize: formatFileSize(
                attachments.reduce((sum, a) => sum + a.size, 0),
              ),
            },
            message: `Found ${formattedAttachments.length} attachment${formattedAttachments.length !== 1 ? "s" : ""} on this story.`,
          };
        }

        case "upload-attachment": {
          if (!storyId || !fileData) {
            return {
              success: false,
              error:
                "Story ID and file data are required for uploading attachments",
            };
          }

          if (userRole === "guest") {
            return {
              success: false,
              error: "Guests cannot upload attachments",
            };
          }

          // Validate file size (10MB limit)
          const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
          if (fileData.size > MAX_FILE_SIZE) {
            return {
              success: false,
              error: `File size (${formatFileSize(fileData.size)}) exceeds the 10MB limit`,
            };
          }

          // Validate file type
          const allowedTypes = [
            "image/jpeg",
            "image/jpg",
            "image/png",
            "image/webp",
            "image/gif",
            "application/pdf",
          ];

          if (!allowedTypes.includes(fileData.type)) {
            return {
              success: false,
              error: `File type ${fileData.type} is not supported`,
            };
          }

          // Convert base64 to File object
          const base64Data = fileData.content.replace(
            /^data:[^;]+;base64,/,
            "",
          );
          const binaryData = Buffer.from(base64Data, "base64");
          const file = new File([binaryData], fileData.name, {
            type: fileData.type,
          });

          const result = await addAttachmentAction(storyId, file);

          if (result.error) {
            return {
              success: false,
              error: result.error.message || "Failed to upload attachment",
            };
          }

          const newAttachment = result.data!;
          const uploader = memberMap.get(newAttachment.uploadedBy);

          return {
            success: true,
            attachment: {
              id: newAttachment.id,
              filename: newAttachment.filename,
              size: newAttachment.size,
              formattedSize: formatFileSize(newAttachment.size),
              mimeType: newAttachment.mimeType,
              fileType: getFileType(newAttachment.mimeType),
              url: newAttachment.url,
              createdAt: newAttachment.createdAt,
              uploader: uploader
                ? {
                    id: uploader.id,
                    name: uploader.fullName,
                    username: uploader.username,
                    avatarUrl: uploader.avatarUrl,
                  }
                : null,
            },
            message: `File "${newAttachment.filename}" uploaded successfully`,
          };
        }

        case "delete-attachment": {
          if (!attachmentId) {
            return {
              success: false,
              error: "Attachment ID is required for deleting attachments",
            };
          }

          // Check if user owns the attachment or is admin
          const attachments = await getStoryAttachments(storyId, session);
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

          const result = await deleteStoryAttachmentAction(
            storyId,
            attachmentId,
          );

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
        }

        default:
          return {
            success: false,
            error: "Invalid attachment action",
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
