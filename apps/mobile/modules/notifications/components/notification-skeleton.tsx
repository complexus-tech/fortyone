import React from "react";
import { View, StyleSheet } from "react-native";

type NotificationSkeletonProps = {
  count?: number;
};

export const NotificationSkeleton = ({ count = 5 }: NotificationSkeletonProps) => {
  const renderSkeletonItem = (index: number) => (
    <View key={index} style={styles.skeletonItem}>
      <View style={styles.skeletonHeader}>
        <View style={styles.skeletonTitle} />
        <View style={styles.skeletonTimestamp} />
      </View>
      
      <View style={styles.skeletonMessageContainer}>
        <View style={styles.skeletonAvatar} />
        <View style={styles.skeletonMessage} />
        <View style={styles.skeletonIcon} />
      </View>
    </View>
  );

  return (
    <View style={styles.container}>
      {Array.from({ length: count }, (_, index) => renderSkeletonItem(index))}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#F2F2F7",
  },
  skeletonItem: {
    backgroundColor: "#FFFFFF",
    borderBottomWidth: 0.5,
    borderBottomColor: "#E5E5EA",
    paddingHorizontal: 16,
    paddingVertical: 12,
  },
  skeletonHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 8,
  },
  skeletonTitle: {
    height: 16,
    backgroundColor: "#E5E5EA",
    borderRadius: 4,
    flex: 1,
    marginRight: 8,
  },
  skeletonTimestamp: {
    height: 14,
    width: 60,
    backgroundColor: "#E5E5EA",
    borderRadius: 4,
  },
  skeletonMessageContainer: {
    flexDirection: "row",
    alignItems: "center",
  },
  skeletonAvatar: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: "#E5E5EA",
    marginRight: 12,
  },
  skeletonMessage: {
    height: 14,
    backgroundColor: "#E5E5EA",
    borderRadius: 4,
    flex: 1,
    marginRight: 8,
  },
  skeletonIcon: {
    width: 16,
    height: 16,
    backgroundColor: "#E5E5EA",
    borderRadius: 2,
  },
});
