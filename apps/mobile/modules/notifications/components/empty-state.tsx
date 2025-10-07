import React from "react";
import { Col, Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { useColorScheme } from "nativewind";

export const EmptyState = () => {
  const { colorScheme } = useColorScheme();
  return (
    <Col justify="center" align="center" className="flex-1" asContainer>
      <Row
        align="center"
        justify="center"
        className="size-18 rounded-full bg-gray-50 mb-6 dark:bg-dark-300"
      >
        <SymbolView
          name="bell.slash.fill"
          size={36}
          tintColor={
            colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
          }
        />
      </Row>
      <Text fontSize="xl" fontWeight="semibold" className="mb-4 text-center">
        No notifications
      </Text>
      <Text color="muted" className="text-center">
        You will receive notifications when you are assigned or mentioned in a
        story.
      </Text>
    </Col>
  );
};
