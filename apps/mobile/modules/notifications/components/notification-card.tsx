import React from "react";
import { View, Text, StyleSheet, Pressable } from "react-native";

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
        <View style={styles.header}>
          <Text
            style={[
              styles.title,
              isUnread ? styles.unreadTitle : styles.readTitle,
            ]}
            numberOfLines={1}
          >
            {notification.title}
          </Text>
          <Text style={styles.timestamp}>2h ago</Text>
        </View>

        <View style={styles.messageContainer}>
          <View style={styles.actorContainer}>
            <View style={styles.avatar}>
              <Text style={styles.avatarText}>JD</Text>
            </View>
            <Text style={styles.message} numberOfLines={2}>
              {notification.message}
            </Text>
          </View>

          <View style={styles.iconContainer}>
            {notification.type === "story_comment" && (
              <Text style={styles.icon}>💬</Text>
            )}
            {notification.type === "mention" && (
              <Text style={styles.icon}>@</Text>
            )}
            {notification.type === "story_update" && (
              <Text style={styles.icon}>📝</Text>
            )}
          </View>
        </View>
      </View>
    </Pressable>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: "#FFFFFF",
    borderBottomWidth: 0.5,
    borderBottomColor: "#E5E5EA",
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  unreadContainer: {
    borderLeftWidth: 3,
    borderLeftColor: "#007AFF",
    backgroundColor: "#F8F9FA",
  },
  pressedContainer: {
    backgroundColor: "#F2F2F7",
  },
  content: {
    flex: 1,
  },
  header: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 8,
  },
  title: {
    fontSize: 16,
    fontWeight: "600",
    flex: 1,
    marginRight: 8,
  },
  unreadTitle: {
    color: "#000000",
  },
  readTitle: {
    color: "#6D6D70",
  },
  timestamp: {
    fontSize: 14,
    color: "#8E8E93",
  },
  messageContainer: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
  },
  actorContainer: {
    flexDirection: "row",
    alignItems: "center",
    flex: 1,
  },
  avatar: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: "#007AFF",
    justifyContent: "center",
    alignItems: "center",
    marginRight: 12,
  },
  avatarText: {
    color: "#FFFFFF",
    fontSize: 14,
    fontWeight: "600",
  },
  message: {
    fontSize: 14,
    color: "#6D6D70",
    flex: 1,
    lineHeight: 20,
  },
  iconContainer: {
    marginLeft: 8,
  },
  icon: {
    fontSize: 16,
  },
});
