import React from "react";
import { View, Text, StyleSheet, Pressable, Platform } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

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
  const insets = useSafeAreaInsets();

  return (
    <View style={[styles.container, { paddingTop: insets.top }]}>
      <View style={styles.blurContainer}>
        <View style={styles.content}>
          <View style={styles.leftSection}>
            <Text style={styles.title}>Inbox</Text>
          </View>

          <View style={styles.rightSection}>
            <Pressable
              style={({ pressed }) => [
                styles.menuButton,
                pressed && styles.pressedButton,
              ]}
              onPress={onFilterPress}
            >
              <Text style={styles.menuDots}>⋯</Text>
            </Pressable>
          </View>
        </View>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: "transparent",
  },
  blurContainer: {
    backgroundColor: "#FFFFFF",
    borderBottomWidth: 1,
    borderBottomColor: "#E5E5EA",
  },
  content: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "flex-start",
    paddingHorizontal: 16,
    paddingTop: 8,
    paddingBottom: 8,
  },
  leftSection: {
    flexDirection: "row",
    alignItems: "center",
  },
  title: {
    fontSize: 32,
    fontWeight: "700",
    color: "#000000",
  },
  rightSection: {
    flexDirection: "row",
    alignItems: "center",
  },
  menuButton: {
    paddingHorizontal: 8,
    paddingVertical: 8,
    borderRadius: 6,
  },
  pressedButton: {
    backgroundColor: "#F2F2F7",
  },
  menuDots: {
    fontSize: 18,
    fontWeight: "600",
    color: "#000000",
  },
});
