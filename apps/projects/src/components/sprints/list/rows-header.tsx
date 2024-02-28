import { Badge, Box, Button, Container, Flex, Text } from "ui";
import { ArrowDownIcon } from "icons";

export const SprintRowsHeader = () => {
  return (
    <Container className="sticky top-0 z-[1] select-none bg-gray-50 py-2.5 backdrop-blur dark:bg-dark-200/60">
      <Flex align="center" justify="between">
        <Flex align="center" gap={2}>
          <Badge color="tertiary" rounded="sm" size="lg">
            Active
          </Badge>
        </Flex>
        <Flex align="center" gap={5}>
          <Text className="w-32 text-left" color="muted">
            Progress
          </Text>
          <Text className="w-32 text-left" color="muted">
            Start date
          </Text>
          <Text className="w-32 text-left" color="muted">
            Target
          </Text>
          <Text className="w-12 text-left" color="muted">
            Lead
          </Text>
          <Text className="w-28 text-left" color="muted">
            Created
          </Text>
          <Box className="w-8">
            <Button
              className="aspect-square"
              color="tertiary"
              rightIcon={<ArrowDownIcon className="h-4 w-auto" />}
              size="sm"
              variant="outline"
            >
              <span className="sr-only">Collapse</span>
            </Button>
          </Box>
        </Flex>
      </Flex>
    </Container>
  );
};
