import { Button, Container, Flex, Text, Tooltip } from "ui";
import { CopyIcon, DeleteIcon, LinkIcon } from "icons";

export const OptionsHeader = () => {
  return (
    <Container className="flex h-16 w-full items-center justify-between px-7">
      <Text color="muted" fontWeight="medium">
        COMP-13
      </Text>
      <Flex gap={2}>
        <Tooltip title="Copy story link">
          <Button
            color="tertiary"
            leftIcon={<LinkIcon className="h-5 w-auto" strokeWidth={2.5} />}
            variant="naked"
          >
            <span className="sr-only">Copy story link</span>
          </Button>
        </Tooltip>
        <Tooltip title="Copy story id">
          <Button
            color="tertiary"
            leftIcon={<CopyIcon className="h-5 w-auto" />}
            variant="naked"
          >
            <span className="sr-only">Copy story id</span>
          </Button>
        </Tooltip>
        <Tooltip title="Delete story">
          <Button
            color="danger"
            leftIcon={<DeleteIcon className="h-5 w-auto" />}
            variant="naked"
          >
            <span className="sr-only">Delete story</span>
          </Button>
        </Tooltip>
      </Flex>
    </Container>
  );
};
