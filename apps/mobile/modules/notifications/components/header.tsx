import React from "react";
import { Alert } from "react-native";
import { Row, Text, ContextMenuButton } from "@/components/ui";

export const Header = () => {
  const handleDeleteRead = () => {
    Alert.alert(
      "Delete read notifications",
      "Are you sure you want to delete all read notifications?",
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Delete",
          style: "destructive",
          onPress: () => {
            console.log("Delete read");
          },
        },
      ]
    );
  };

  const handleDeleteAll = () => {
    Alert.alert(
      "Delete all notifications",
      "Are you sure you want to delete all notifications?",
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Delete",
          style: "destructive",
          onPress: () => {
            console.log("Delete all");
          },
        },
      ]
    );
  };

  const handleMarkAllAsRead = () => {
    console.log("Mark all as read");
  };

  return (
    <Row justify="between" align="center" asContainer className="mb-2">
      <Text fontSize="2xl" fontWeight="semibold" color="black">
        Notifications
      </Text>
      <ContextMenuButton
        actions={[
          {
            systemImage: "checkmark.circle.fill",
            label: "Mark all as read",
            onPress: () => {},
          },
          {
            systemImage: "delete.forward.fill",
            label: "Delete read",
            onPress: handleDeleteRead,
          },
          {
            systemImage: "trash.fill",
            label: "Delete all",
            onPress: handleDeleteAll,
          },
        ]}
      />
    </Row>
  );
};
