import React from "react";
import { View, Text, StyleSheet } from "react-native";

interface StatsCardProps {
  title: string;
  count: number;
}

export const StatsCard = ({ title, count }: StatsCardProps) => {
  return (
    <View style={styles.card}>
      <Text style={styles.count}>{count}</Text>
      <Text style={styles.title}>{title}</Text>
    </View>
  );
};

const styles = StyleSheet.create({
  card: {
    backgroundColor: "#FFFFFF",
    borderRadius: 12,
    padding: 16,
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
    alignItems: "center",
    minHeight: 80,
  },
  count: {
    fontSize: 24,
    fontWeight: "600",
    color: "#000000",
    marginBottom: 8,
  },
  title: {
    fontSize: 12,
    color: "#666",
    textAlign: "center",
    lineHeight: 16,
  },
});
