import React from "react";
import { View, Pressable, ActivityIndicator } from "react-native";
import { Text, Row } from "@/components/ui";
import { colors } from "@/constants";

type SectionFooterProps = {
  hasMore: boolean;
  loadedCount: number;
  totalCount: number;
  isLoading?: boolean;
  onLoadMore: () => void;
};

export const SectionFooter = ({
  hasMore,
  loadedCount,
  totalCount,
  isLoading = false,
  onLoadMore,
}: SectionFooterProps) => {
  if (!hasMore) {
    return (
      <View
        style={{
          paddingVertical: 12,
          paddingHorizontal: 16,
        }}
      >
        <Text fontSize="sm" color="muted">
          Showing {loadedCount} {loadedCount === 1 ? "story" : "stories"}
        </Text>
      </View>
    );
  }

  return (
    <View
      style={{
        paddingVertical: 8,
        paddingHorizontal: 16,
      }}
    >
      <Pressable
        onPress={onLoadMore}
        disabled={isLoading}
        style={({ pressed }) => [
          {
            paddingVertical: 8,
            paddingHorizontal: 12,
            borderRadius: 6,
            backgroundColor: pressed ? colors.gray[100] : colors.gray[50],
          },
        ]}
      >
        <Row align="center" justify="between">
          {isLoading ? (
            <Row align="center" gap={2}>
              <ActivityIndicator size="small" color={colors.gray.DEFAULT} />
              <Text fontSize="sm" color="muted">
                Loading...
              </Text>
            </Row>
          ) : (
            <>
              <Text fontSize="sm" fontWeight="medium" color="primary">
                Load more stories
              </Text>
              <Text fontSize="sm" color="muted">
                {loadedCount} of {totalCount}
              </Text>
            </>
          )}
        </Row>
      </Pressable>
    </View>
  );
};
