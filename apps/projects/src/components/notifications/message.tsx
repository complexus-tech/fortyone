import { NotificationsIcon } from "icons";
import { Flex, Text } from "ui";

export const NotificationMessage = () => {
  return (
    <Flex align="center" className="h-full" justify="center">
      <Flex align="center" direction="column">
        <Text className="mb-5">Select a notification to read</Text>
        <NotificationsIcon
          className="mb-3 h-28 w-auto dark:text-gray"
          strokeWidth={1}
        />
        <Text color="muted">Nothing is selected.</Text>
      </Flex>
    </Flex>
  );
};
