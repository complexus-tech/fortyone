import { Container, Flex, Text } from "ui";

export const Heading = () => {
  return (
    <Container className="sticky top-0 z-[1] h-[3.2rem] border-b-[0.5px] border-gray-100/60 bg-gray-50/60 leading-[3.2rem] backdrop-blur dark:border-dark-100 dark:bg-dark-300/90">
      <Flex align="center" justify="between">
        <Text color="muted" fontWeight="medium">
          Name
        </Text>
        <Flex align="center" gap={5}>
          <Text className="w-40 text-left" color="muted">
            Progress
          </Text>
          <Text className="w-40 text-left" color="muted">
            Start date
          </Text>
          <Text className="w-40 text-left" color="muted">
            Target
          </Text>
          <Text className="w-40 text-left" color="muted">
            Lead
          </Text>
          <Text className="w-40 text-left" color="muted">
            Created
          </Text>
        </Flex>
      </Flex>
    </Container>
  );
};
