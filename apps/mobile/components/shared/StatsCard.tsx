import React from "react";
import { View, Text, StyleSheet, Pressable } from "react-native";
import { Ionicons } from "@expo/vector-icons";

interface StatsCardProps {
  title: string;
  count: number;
  onPress?: () => void;
}

export const StatsCard = ({ title, count, onPress }: StatsCardProps) => {
  return (
    <Pressable
      style={({ pressed }) => [styles.card, pressed && styles.pressedCard]}
      onPress={onPress}
    >
      <View style={styles.header}>
        <Text style={styles.count}>{count}</Text>
        <Ionicons name="chevron-forward" size={16} color="#C7C7CC" />
      </View>
      <Text style={styles.title}>{title}</Text>
    </Pressable>
  );
};

const styles = StyleSheet.create({
  card: {
    backgroundColor: "#FFFFFF",
    borderRadius: 12,
    paddingHorizontal: 10,
    paddingVertical: 16,
    marginHorizontal: 4,
    marginVertical: 4,
    shadowColor: "#000",
    shadowOffset: {
      width: 0,
      height: 1,
    },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 2,
    flex: 1,
  },
  pressedCard: {
    backgroundColor: "#F2F2F7",
  },
  header: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    width: "100%",
    marginBottom: 8,
  },
  count: {
    fontSize: 24,
    fontWeight: "600",
    color: "#000000",
  },
  title: {
    fontSize: 12,
    color: "#666",
    lineHeight: 16,
  },
});
