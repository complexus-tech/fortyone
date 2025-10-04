import React from "react";
import { Col, Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

type EmptyStateProps = {
  title?: string;
  message?: string;
};

export const EmptyState = ({
  title = "No sprints found",
  message = "There are no sprints to display at the moment.",
}: EmptyStateProps) => {
  return (
    <Col justify="center" align="center" className="flex-1" asContainer>
      <Row
        align="center"
        justify="center"
        className="size-18 rounded-full bg-gray-50 mb-6"
      >
        <SymbolView name="memories" size={36} tintColor={colors.gray.DEFAULT} />
      </Row>
      <Text fontSize="xl" fontWeight="semibold" className="mb-4 text-center">
        {title}
      </Text>
      <Text color="muted" className="text-center">
        {message}
      </Text>
    </Col>
  );
};
