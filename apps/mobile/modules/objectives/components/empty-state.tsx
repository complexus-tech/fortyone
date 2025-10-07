import React from "react";
import { Col, Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { useTerminology } from "@/hooks/use-terminology";
import { useColorScheme } from "nativewind";

type EmptyStateProps = {
  title?: string;
  message?: string;
};

export const EmptyState = ({ title, message }: EmptyStateProps) => {
  const { colorScheme } = useColorScheme();
  const { getTermDisplay } = useTerminology();

  const defaultTitle = `No ${getTermDisplay("objectiveTerm", { variant: "plural" })} found`;
  const defaultMessage = `There are no ${getTermDisplay("objectiveTerm", { variant: "plural" })} to display at the moment.`;

  return (
    <Col justify="center" align="center" className="flex-1 pt-56" asContainer>
      <Row
        align="center"
        justify="center"
        className="size-18 rounded-full bg-gray-50 mb-6 dark:bg-dark-300"
      >
        <SymbolView
          name="trash"
          size={36}
          tintColor={
            colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[200]
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
