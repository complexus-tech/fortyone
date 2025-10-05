import React from "react";
import { Pressable, ActivityIndicator } from "react-native";
import { Text, Row } from "@/components/ui";
import { colors } from "@/constants";
import { useTerminology } from "@/hooks/use-terminology";

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
  const { getTermDisplay } = useTerminology();

  if (!hasMore) {
    return null;
  }

  return (
    <Row asContainer className="py-3">
      <Pressable className="flex-1" onPress={onLoadMore} disabled={isLoading}>
        {isLoading ? (
          <Row align="center" gap={2}>
            <ActivityIndicator size="small" color={colors.gray.DEFAULT} />
            <Text fontSize="sm" color="muted">
              Loading...
            </Text>
          </Row>
        ) : (
          <Row align="center" gap={2} justify="between">
            <Text fontSize="sm" fontWeight="semibold">
              Load more {getTermDisplay("storyTerm", { variant: "plural" })}
            </Text>
            <Text fontSize="sm" color="muted">
              {loadedCount} of {totalCount}
            </Text>
          </Row>
        )}
      </Pressable>
    </Row>
  );
};
