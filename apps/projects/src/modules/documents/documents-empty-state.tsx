import { Box, Flex, Text } from "ui";
import { DocumentsEmptyIllustration } from "@/components/ui/illustrations/empty-state-illustrations";

export const DocumentsEmptyState = () => (
  <Flex align="center" className="h-full px-8" justify="center">
    <Box className="max-w-sm text-center">
      <DocumentsEmptyIllustration className="mx-auto mb-5 w-52" />
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
