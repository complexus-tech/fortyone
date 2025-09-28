import React from "react";
import { View, Text, StyleSheet, Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";

interface TeamLinkProps {
  name: string;
  color: string;
  onPress?: () => void;
}

export const TeamLink = ({ name, color, onPress }: TeamLinkProps) => {
  return (
    <Pressable
      style={({ pressed }) => [styles.teamLink, pressed && styles.pressedItem]}
      onPress={onPress}
    >
      <View style={styles.itemContent}>
        <View style={styles.leftContent}>
          <View style={[styles.colorIndicator, { backgroundColor: color }]} />
          <Text style={styles.teamName}>{name}</Text>
        </View>
        <Ionicons name="chevron-forward" size={16} color="#C7C7CC" />
      </View>
    </Pressable>
  );
};

const styles = StyleSheet.create({
  teamLink: {
    backgroundColor: "#FFFFFF",
  },
  pressedItem: {
    backgroundColor: "#F2F2F7",
  },
  itemContent: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 16,
    paddingVertical: 14,
    minHeight: 44,
  },
  leftContent: {
    flexDirection: "row",
    alignItems: "center",
  },
  colorIndicator: {
    width: 12,
    height: 12,
    borderRadius: 5,
    marginRight: 8,
  },
  teamName: {
    fontSize: 16,
    color: "#000000",
  },
});
