import { NotificationsIcon } from "icons";
import { Flex, Text } from "ui";

export const NotificationMessage = ({ count }: { count: number }) => {
  return (
    <Flex align="center" className="h-full" justify="center">
      <Flex align="center" direction="column">
        <NotificationsIcon className="mb-3 h-16 w-auto" />
        {count === 0 ? (
          <Text color="muted">You have no notifications</Text>
        ) : (
          <Text color="muted">Select a notification to view</Text>
        )}
      </Flex>
    </Flex>
  );
};
