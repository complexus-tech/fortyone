import { Box, Text } from "ui";
import { AttachmentIcon } from "icons";

export const Attachments = () => {
  return (
    <Box>
      <Text as="h4" className="flex items-center gap-1" fontWeight="medium">
        <AttachmentIcon className="h-6 w-auto" />
        Attachements
      </Text>
      <Box className="mb-4 mt-3 flex h-24 cursor-pointer items-center justify-center rounded-xl border-[1.5px] border-dashed border-gray-200 bg-gray-50/50 dark:border-dark-100 dark:bg-dark-200/40">
        <Text color="muted">Click or drag files here</Text>
      </Box>

      <Box className="grid grid-cols-5 gap-4">
        <Box className="h-28 rounded-xl bg-gray-50/70 dark:bg-dark-200/50" />
        <Box className="h-28 rounded-xl bg-gray-50/70 dark:bg-dark-200/50" />
        <Box className="h-28 rounded-xl bg-gray-50/70 dark:bg-dark-200/50" />

        <Box className="h-28 rounded-xl bg-gray-50/70 dark:bg-dark-200/50" />
        <Box className="h-28 rounded-xl bg-gray-50/70 dark:bg-dark-200/50" />
        <Box className="h-28 rounded-xl bg-gray-50/70 dark:bg-dark-200/50" />
      </Box>
    </Box>
  );
};
