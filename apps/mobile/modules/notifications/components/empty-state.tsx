import React from "react";
import { Col, Row, Text } from "@/components/ui";
import { Ionicons } from "@expo/vector-icons";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";

export const EmptyState = () => {
  const { resolvedTheme } = useTheme();
  return (
    <Col justify="center" align="center" className="flex-1 px-[20px] pb-16">
      <Row
        align="center"
        justify="center"
        className="size-[48px] rounded-2xl bg-gray-50 mb-4 dark:bg-dark-200"
      >
        <Ionicons
          name="checkmark-done-outline"
          size={25}
          color={themeColors[resolvedTheme].textMuted}
        />
      </Row>
      <Text fontSize="xl" fontWeight="semibold" className="mb-2 text-center">
        You’re all caught up
      </Text>
      <Text color="muted" className="text-center">
        Updates, assignments, and mentions will appear here.
      </Text>
    </Col>
  );
};
