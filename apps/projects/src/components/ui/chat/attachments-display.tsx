import { Box, Flex } from "ui";
import { StoryAttachmentPreview } from "@/modules/story/components/story-attachment-preview";
import type { MayaUIMessage } from "@/lib/ai/tools/types";

export const AttachmentsDisplay = ({ message }: { message: MayaUIMessage }) => {
  return (
    <Box className="w-full">
      <Flex className="items-end gap-2" justify="end" wrap>
        {message.parts.map((part, idx) => {
          if (part.type === "file" && part.mediaType.startsWith("/image")) {
            return (
              <StoryAttachmentPreview
                className="w-[8.2rem]"
                file={{
                  id: part.filename + idx.toString(),
                  filename: part.filename ?? "Attachment",
                  size: 0,
                  mimeType: part.mediaType,
                  url: part.url,
                  createdAt: new Date().toISOString(),
                  uploadedBy: "me",
                }}
                isInChat
                key={idx}
              />
            );
          }
          return null;
        })}
      </Flex>
      <Flex direction="column" gap={2}>
        {message.parts.map((part, idx) => {
          if (part.type === "file" && part.mediaType.includes("pdf")) {
            return (
              <StoryAttachmentPreview
                file={{
                  id: part.filename + idx.toString(),
                  filename: part.filename ?? "Attachment",
                  size: 0,
                  mimeType: part.mediaType,
                  url: part.url,
                  createdAt: new Date().toISOString(),
                  uploadedBy: "me",
                }}
                isInChat
                key={idx}
              />
            );
          }
          return null;
        })}
      </Flex>
    </Box>
  );
};
