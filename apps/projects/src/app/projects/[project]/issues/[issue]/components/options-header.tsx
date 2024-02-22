import { Box, Button, Container, Flex, Text, Tooltip } from "ui";
import { CopyIcon, DeleteIcon, LinkIcon } from "icons";

export const OptionsHeader = () => {
  return (
    <Box className="flex h-16 items-center border-b border-gray-50 dark:border-dark-200">
      <Container className="flex w-full items-center justify-between px-8">
        <Text color="muted" fontWeight="medium">
          COMP-13
        </Text>
        <Flex gap={2}>
          <Tooltip title="Copy issue link">
            <Button
              color="tertiary"
              leftIcon={<LinkIcon className="h-5 w-auto" strokeWidth={2.5} />}
              variant="naked"
            >
              <span className="sr-only">Copy issue link</span>
            </Button>
          </Tooltip>
          <Tooltip title="Copy issue id">
            <Button
              color="tertiary"
              leftIcon={<CopyIcon className="h-5 w-auto" />}
              variant="naked"
            >
              <span className="sr-only">Copy issue id</span>
            </Button>
          </Tooltip>
          <Tooltip title="Delete issue">
            <Button
              color="danger"
              leftIcon={<DeleteIcon className="h-5 w-auto" />}
              variant="naked"
            >
              <span className="sr-only">Delete issue</span>
            </Button>
          </Tooltip>
        </Flex>
      </Container>
    </Box>
  );
};
