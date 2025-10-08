import React from "react";
import { Col, Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

type EmptyStateProps = {
  title?: string;
  message?: string;
};

export const EmptyState = ({ title, message }: EmptyStateProps) => {
  const defaultTitle = "No links found";
  const defaultMessage = "This story doesn't have any links yet.";

  return (
    <Col justify="center" align="center" className="flex-1 mt-52" asContainer>
      <Row
        align="center"
        justify="center"
        className="size-18 rounded-full bg-gray-50 dark:bg-dark-200 mb-6"
      >
        <SymbolView name="link" size={36} tintColor={colors.gray.DEFAULT} />
      </Row>
      <Text fontSize="xl" fontWeight="semibold" className="mb-4 text-center">
        {title || defaultTitle}
      </Text>
      <Text color="muted" className="text-center">
        {message || defaultMessage}
      </Text>
    </Col>
  );
};
