import { Inbox } from "lucide-react";
import { Flex, Text } from "ui";

export const Message = () => {
  return (
    <Flex align="center" className="h-full" justify="center">
      <Flex align="center" direction="column">
        <Inbox className="h-28 w-auto dark:text-gray" strokeWidth={0.8} />
        <Text color="muted">Select a notification to read.</Text>
      </Flex>
    </Flex>
  );
};
