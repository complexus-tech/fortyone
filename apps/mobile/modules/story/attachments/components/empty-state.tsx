import React from "react";
import { Col, Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import { useColorScheme } from "nativewind";

type EmptyStateProps = {
  title?: string;
  message?: string;
};

export const EmptyState = ({ title, message }: EmptyStateProps) => {
  const { colorScheme } = useColorScheme();

  const defaultTitle = "No attachments found";
  const defaultMessage = "There are no attachments to display at the moment.";

  return (
    <Col justify="center" align="center" className="flex-1 mt-52" asContainer>
      <Row
        align="center"
        justify="center"
        className="size-18 rounded-full bg-gray-50 dark:bg-dark-200 mb-6"
      >
        <SymbolView
          name="paperclip"
          size={36}
          tintColor={
            colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
          }
        />
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
