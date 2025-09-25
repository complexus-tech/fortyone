import React from "react";
import { View, Text, StyleSheet, Pressable } from "react-native";

type NotificationHeaderProps = {
  unreadCount?: number;
  onFilterPress?: () => void;
  onMarkAllRead?: () => void;
  onDeleteAll?: () => void;
};

export const NotificationHeader = ({
  unreadCount = 0,
  onFilterPress,
  onMarkAllRead,
  onDeleteAll,
}: NotificationHeaderProps) => {
  return (
    <View style={styles.container}>
      <View style={styles.leftSection}>
        <Text style={styles.title}>Notifications</Text>
        {unreadCount > 0 && (
          <View style={styles.badge}>
            <Text style={styles.badgeText}>
              {unreadCount > 99 ? "99+" : unreadCount}
            </Text>
          </View>
        )}
      </View>

      <View style={styles.rightSection}>
        <Pressable
          style={({ pressed }) => [
            styles.actionButton,
            pressed && styles.pressedButton,
          ]}
          onPress={onFilterPress}
        >
          <Text style={styles.actionText}>Filter</Text>
        </Pressable>

        <Pressable
          style={({ pressed }) => [
            styles.actionButton,
            pressed && styles.pressedButton,
          ]}
          onPress={onMarkAllRead}
        >
          <Text style={styles.actionText}>Mark All</Text>
        </Pressable>

        <Pressable
          style={({ pressed }) => [
            styles.actionButton,
            pressed && styles.pressedButton,
          ]}
          onPress={onDeleteAll}
        >
          <Text style={styles.actionText}>Delete</Text>
        </Pressable>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    paddingHorizontal: 16,
    paddingVertical: 12,
    backgroundColor: "#FFFFFF",
    borderBottomWidth: 0.5,
    borderBottomColor: "#E5E5EA",
  },
  leftSection: {
    flexDirection: "row",
    alignItems: "center",
  },
  title: {
    fontSize: 20,
    fontWeight: "700",
    color: "#000000",
  },
  badge: {
    backgroundColor: "#FF3B30",
    borderRadius: 10,
    paddingHorizontal: 6,
    paddingVertical: 2,
    marginLeft: 8,
    minWidth: 20,
    alignItems: "center",
  },
  badgeText: {
    color: "#FFFFFF",
    fontSize: 12,
    fontWeight: "600",
  },
  rightSection: {
    flexDirection: "row",
    gap: 8,
  },
  actionButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 6,
    backgroundColor: "#F2F2F7",
  },
  pressedButton: {
    backgroundColor: "#E5E5EA",
  },
  actionText: {
    fontSize: 14,
    fontWeight: "500",
    color: "#007AFF",
  },
});
