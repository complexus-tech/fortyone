import { ChatIcon } from "icons";
import { Flex, Text } from "ui";

export const PublicPortalNotFoundState = ({
  description,
  title,
}: {
  description: string;
  title: string;
}) => (
  <Flex
    align="center"
    className="bg-background min-h-dvh px-6"
    justify="center"
  >
    <Flex align="center" className="max-w-md text-center" direction="column">
      <Flex
        align="center"
        className="bg-surface-muted text-text-muted size-12 rounded-xl"
        justify="center"
      >
        <ChatIcon className="h-5 w-auto" />
      </Flex>
      <Text as="h1" className="mt-5 text-2xl font-semibold">
        {title}
      </Text>
      <Text className="mt-2" color="muted">
        {description}
      </Text>
    </Flex>
  </Flex>
);
