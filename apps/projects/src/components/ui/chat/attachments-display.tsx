import { Box, Flex } from "ui";
import { StoryAttachmentPreview } from "@/modules/story/components/story-attachment-preview";
import type { MayaUIMessage } from "@/lib/ai/tools/types";

const getFilePartKey = ({
  filename,
  mediaType,
  url,
}: {
  filename?: string;
  mediaType: string;
  url: string;
}) =>
  `${filename ?? "attachment"}-${mediaType}-${url.length}-${url.slice(-32)}`;

export const AttachmentsDisplay = ({ message }: { message: MayaUIMessage }) => {
  const fileParts = message.parts.filter((part) => part.type === "file");
  const imageParts = fileParts.filter((part) =>
    part.mediaType.startsWith("image/"),
  );
  const pdfParts = fileParts.filter(
    (part) => part.mediaType === "application/pdf",
  );

  return (
    <Box className="w-full">
      <Flex className="items-end gap-2" justify="end" wrap>
        {imageParts.map((part) => (
          <StoryAttachmentPreview
            className="w-[8.2rem]"
            file={{
              id: getFilePartKey(part),
              filename: part.filename ?? "Attachment",
              size: 0,
              mimeType: part.mediaType,
              url: part.url,
              createdAt: "",
              uploadedBy: "me",
            }}
            isInChat
            key={getFilePartKey(part)}
          />
        ))}
      </Flex>
      <Flex direction="column" gap={2}>
        {pdfParts.map((part) => (
          <StoryAttachmentPreview
            file={{
              id: getFilePartKey(part),
              filename: part.filename ?? "Attachment",
              size: 0,
              mimeType: part.mediaType,
              url: part.url,
              createdAt: "",
              uploadedBy: "me",
            }}
            isInChat
            key={getFilePartKey(part)}
          />
        ))}
      </Flex>
    </Box>
  );
};
