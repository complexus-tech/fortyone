import React from "react";
import { View, Text, StyleSheet } from "react-native";

interface SectionProps {
  title: string;
  children: React.ReactNode;
  style?: any;
}

export const Section = ({ title, children, style }: SectionProps) => {
  return (
    <View style={[styles.section, style]}>
      <Text style={styles.sectionTitle}>{title}</Text>
      <View style={styles.sectionContent}>{children}</View>
    </View>
  );
};

const styles = StyleSheet.create({
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: "400",
    color: "#8E8E93",
    textTransform: "uppercase",
    letterSpacing: 0.5,
    marginBottom: 8,
    marginHorizontal: 16,
  },
  sectionContent: {
    backgroundColor: "#FFFFFF",
  },
});
