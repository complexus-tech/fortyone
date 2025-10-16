import React from "react";
import { ScrollView, RefreshControl } from "react-native";

type StoriesListProps = {
  children: React.ReactNode;
  onRefresh?: () => void;
  isRefreshing?: boolean;
};

export const StoriesList = ({
  children,
  onRefresh,
  isRefreshing = false,
}: StoriesListProps) => {
  return (
    <ScrollView
      refreshControl={
        onRefresh ? (
          <RefreshControl
            refreshing={isRefreshing}
            onRefresh={onRefresh}
            tintColor="#007AFF"
            colors={["#007AFF"]}
          />
        ) : undefined
      }
      showsVerticalScrollIndicator={false}
      contentContainerStyle={{ paddingBottom: 120 }}
    >
      {children}
    </ScrollView>
  );
};
