import React from "react";
import { View, Text, StyleSheet, Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";

interface CycleItemProps {
  title: string;
  status: "in-review" | "in-progress" | "todo";
  assignee?: {
    name: string;
    avatar?: string;
  };
  onPress?: () => void;
}

const statusConfig = {
  "in-review": {
    icon: "checkmark-circle",
    color: "#34C759",
    label: "In Review",
  },
  "in-progress": {
    icon: "time",
    color: "#666",
    label: "In Progress",
  },
  todo: {
    icon: "radio-button-off",
    color: "#666",
    label: "Todo",
  },
};

export const CycleItem = ({
  title,
  status,
  assignee,
  onPress,
}: CycleItemProps) => {
  const config = statusConfig[status];

  const getStatusIcon = () => {
    if (status === "in-review") return "✓";
    if (status === "in-progress") return "→";
    return "○";
  };

  const getStatusColor = () => {
    if (status === "in-review") return "#34C759";
    if (status === "in-progress") return "#FF9500";
    return "#8E8E93";
  };

  return (
    <Pressable
      style={({ pressed }) => [
        styles.container,
        pressed && styles.pressedContainer,
      ]}
      onPress={onPress}
    >
      <View style={styles.content}>
        <View style={styles.leftSection}>
          <View style={styles.iconContainer}>
            <View style={styles.avatar}>
              <Text style={styles.avatarText}>
                {assignee ? assignee.name.charAt(0).toUpperCase() : "U"}
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
              {title}
            </Text>
            <Text style={styles.timestamp}>2mo ago</Text>
          </View>

          {assignee && (
            <Text style={styles.message} numberOfLines={1}>
              {assignee.name}
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
