import React from "react";
import { Row, Text, ContextMenuButton } from "@/components/ui";

export const Header = () => {
  return (
    <Row justify="between" align="center" asContainer className="mb-2">
      <Text fontSize="2xl" fontWeight="semibold" color="black">
        Notifications
      </Text>

      <ContextMenuButton
        actions={[
          {
            systemImage: "gear",
            label: "Notification settings",
            onPress: () => {},
          },
          {
            systemImage: "checkmark.circle.fill",
            label: "Mark all as read",
            onPress: () => {},
          },
          {
            systemImage: "delete.forward.fill",
            label: "Delete read",
            onPress: () => {},
          },
          {
            systemImage: "trash.fill",
            label: "Delete all",
            onPress: () => {},
          },
        ]}
      />
    </Row>
  );
};
