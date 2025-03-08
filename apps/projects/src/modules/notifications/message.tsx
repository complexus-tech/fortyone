import { NotificationsIcon } from "icons";
import { Flex, Text } from "ui";

export const NotificationMessage = () => {
  return (
    <Flex align="center" className="h-full" justify="center">
      <Flex align="center" direction="column">
        <NotificationsIcon className="mb-3 h-20 w-auto" />
        <Text color="muted">You have no notifications</Text>
      </Flex>
    </Flex>
  );
};
