import React from "react";
import { Alert } from "react-native";
import { ScreenHeader, HeaderActions } from "@/components/ui";
import { useRouter } from "expo-router";
import { useTerminology } from "@/hooks/use-terminology";
import { colors } from "@/constants";
import {
  useDeleteAllMutation,
  useDeleteReadMutation,
  useReadAllNotificationsMutation,
} from "../hooks";

export const Header = () => {
  const router = useRouter();
  const { getTermDisplay } = useTerminology();
  const readAllMutation = useReadAllNotificationsMutation();
  const { mutate: deleteRead } = useDeleteReadMutation();
  const { mutate: deleteAll } = useDeleteAllMutation();

  const handleMarkAllAsRead = () => {
    readAllMutation.mutate();
  };

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
            deleteRead();
          },
        },
      ],
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
            deleteAll();
          },
        },
      ],
    );
  };

  return (
    <ScreenHeader
      title="Inbox"
      trailing={
        <HeaderActions
          createLabel={`Create ${getTermDisplay("storyTerm")}`}
          onCreate={() => router.push("/new")}
          actions={[
            {
              systemImage: "checkmark.circle.fill",
              label: "Mark all as read",
              onPress: handleMarkAllAsRead,
            },
            {
              systemImage: "delete.forward.fill",
              label: "Delete read",
              onPress: handleDeleteRead,
              color: colors.danger,
            },
            {
              systemImage: "trash.fill",
              label: "Delete all",
              onPress: handleDeleteAll,
              color: colors.danger,
            },
          ]}
        />
      }
    />
  );
};
