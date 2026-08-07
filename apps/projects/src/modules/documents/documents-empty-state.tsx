import { Box, Flex, Text } from "ui";
import { DocsIcon } from "icons";

export const DocumentsEmptyState = () => (
  <Flex align="center" className="h-dvh px-8" justify="center">
    <Box className="max-w-sm text-center">
      <Flex
        align="center"
        className="bg-surface-muted text-text-muted mx-auto mb-5 size-12 rounded-2xl"
        justify="center"
      >
        <DocsIcon className="size-6" />
      </Flex>
      <Text className="mb-2" fontSize="xl" fontWeight="semibold">
        Select a document
      </Text>
      <Text color="muted">
        Open a document from the sidebar or create one to capture plans,
        decisions, and context alongside your work.
      </Text>
    </Box>
  </Flex>
);
