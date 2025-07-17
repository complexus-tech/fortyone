import { Box, Text, Flex } from "ui";
import { getFileType } from "@/lib/utils/files";

interface AttachmentDisplayProps {
  attachment: {
    name: string;
    contentType: string;
    url: string;
  };
}

export const AttachmentDisplay = ({ attachment }: AttachmentDisplayProps) => {
  // Create a temporary File object for our utility functions
  const tempFile = {
    name: attachment.name,
    type: attachment.contentType,
    size: 0, // We don't have size info from the attachment data
  } as File;

  const fileType = getFileType(tempFile);

  return (
    <Box className="mt-2 inline-block rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-dark-100 dark:bg-dark-200">
      <Flex align="center" gap={3}>
        <Box className="bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400 flex h-8 w-8 items-center justify-center rounded text-xs font-medium">
          {fileType === "image" ? "IMG" : "PDF"}
        </Box>
        <Box>
          <Text className="text-sm font-medium">{attachment.name}</Text>
          <Text className="text-gray-500 dark:text-gray-400 text-xs">
            {fileType}
          </Text>
        </Box>
      </Flex>
    </Box>
  );
};
