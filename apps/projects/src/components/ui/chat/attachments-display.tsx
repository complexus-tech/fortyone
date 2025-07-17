import type { Attachment } from "ai";
import { Box, Flex } from "ui";
import { StoryAttachmentPreview } from "@/modules/story/components/story-attachment-preview";

export const AttachmentsDisplay = ({
  attachments,
}: {
  attachments?: Attachment[];
}) => {
  const images =
    attachments?.filter((attachment) =>
      attachment.contentType?.startsWith("image/"),
    ) ?? [];
  const pdfs =
    attachments?.filter((attachment) =>
      attachment.contentType?.includes("pdf"),
    ) ?? [];

  return (
    <Box className="w-full">
      {images.length > 0 && (
        <Flex className="mt-2.5 items-end gap-2" justify="end" wrap>
          {images.map((attachment, idx) => (
            <StoryAttachmentPreview
              className="w-[8.2rem]"
              file={{
                id: attachment.name ?? idx.toString(),
                filename: attachment.name ?? "Attachment",
                size: 0,
                mimeType: attachment.contentType ?? "",
                url: attachment.url,
                createdAt: new Date().toISOString(),
                uploadedBy: "me",
              }}
              isInChat
              key={idx}
            />
          ))}
        </Flex>
      )}
      {pdfs.length > 0 && (
        <Flex className="mt-2.5" direction="column" gap={2}>
          {pdfs.map((attachment, idx) => (
            <StoryAttachmentPreview
              file={{
                id: attachment.name ?? idx.toString(),
                filename: attachment.name ?? "Attachment",
                size: 0,
                mimeType: attachment.contentType ?? "",
                url: attachment.url,
                createdAt: new Date().toISOString(),
                uploadedBy: "me",
              }}
              isInChat
              key={idx}
            />
          ))}
        </Flex>
      )}
    </Box>
  );
};
