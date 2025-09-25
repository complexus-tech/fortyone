import React from "react";
import { View, Text, StyleSheet, Pressable, Platform } from "react-native";

type NotificationCardProps = {
  notification: {
    id: string;
    title: string;
    message: string;
    type: "story_update" | "story_comment" | "mention";
    actor: {
      name: string;
      avatar?: string;
    };
    createdAt: string;
    readAt: string | null;
    entityId: string;
    repository?: string;
    issueNumber?: number;
    status?: string;
  };
  index: number;
  onPress?: () => void;
  onLongPress?: () => void;
};

export const NotificationCard = ({
  notification,
  onPress,
  onLongPress,
}: NotificationCardProps) => {
  const isUnread = !notification.readAt;

  const getStatusIcon = () => {
    if (notification.status === "closed") return "✕";
    if (notification.status === "merged") return "✓";
    if (notification.status === "urgent") return "!";
    return "→";
  };

  const getStatusColor = () => {
    if (notification.status === "closed") return "#6D6D70";
    if (notification.status === "urgent") return "#FF9500";
    return "#6D6D70";
  };

  return (
    <Pressable
      style={({ pressed }) => [
        styles.container,
        isUnread && styles.unreadContainer,
        pressed && styles.pressedContainer,
      ]}
      onPress={onPress}
      onLongPress={onLongPress}
    >
      <View style={styles.content}>
        <View style={styles.leftSection}>
          <View style={styles.iconContainer}>
            <View style={styles.avatar}>
              <Text style={styles.avatarText}>
                {notification.actor.name.charAt(0).toUpperCase()}
              </Text>
            </View>
            <View
              style={[styles.statusIcon, { backgroundColor: getStatusColor() }]}
            >
              <Text style={styles.statusIconText}>{getStatusIcon()}</Text>
            </View>
          </View>
        </View>

        <View style={styles.rightSection}>
          <View style={styles.titleRow}>
            <Text style={styles.title} numberOfLines={1}>
              {notification.title}
            </Text>
            <Text style={styles.timestamp}>2mo ago</Text>
          </View>

          {notification.message && (
            <Text style={styles.message} numberOfLines={1}>
              {notification.message}
            </Text>
          )}
        </View>
      </View>
    </Pressable>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: "#FFFFFF",
    borderBottomWidth: 1,
    borderBottomColor: "#F2F2F7",
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  unreadContainer: {
    backgroundColor: "#F8F9FA",
  },
  pressedContainer: {
    backgroundColor: "#F2F2F7",
  },
  content: {
    flexDirection: "row",
  },
  leftSection: {
    marginRight: 12,
    alignItems: "center",
  },
  iconContainer: {
    alignItems: "center",
    justifyContent: "center",
    position: "relative",
  },
  avatar: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: "#000000",
    justifyContent: "center",
    alignItems: "center",
  },
  avatarText: {
    color: "#FFFFFF",
    fontSize: 16,
    fontWeight: "600",
  },
  statusIcon: {
    position: "absolute",
    bottom: -2,
    right: -2,
    width: 16,
    height: 16,
    borderRadius: 8,
    justifyContent: "center",
    alignItems: "center",
  },
  statusIconText: {
    color: "#FFFFFF",
    fontSize: 10,
    fontWeight: "600",
  },
  rightSection: {
    flex: 1,
  },
  titleRow: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 2,
  },
  title: {
    fontSize: 16,
    fontWeight: "600",
    color: "#000000",
    flex: 1,
    marginRight: 8,
  },
  message: {
    fontSize: 14,
    color: "#6D6D70",
    lineHeight: 20,
  },
  timestamp: {
    fontSize: 14,
    color: "#6D6D70",
  },
});
