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
  const iconColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const defaultTitle = `No sub ${getTermDisplay("storyTerm", { variant: "plural" })} found`;
  const defaultMessage = `This ${getTermDisplay("storyTerm", { variant: "singular" })} doesn't have any sub ${getTermDisplay("storyTerm", { variant: "plural" })} yet.`;

  return (
    <Col justify="center" align="center" className="flex-1 mt-52" asContainer>
      <Row
        align="center"
        justify="center"
        className="size-18 rounded-full bg-gray-50 dark:bg-dark-300 mb-6"
      >
        <SymbolView
          name="checklist.unchecked"
          size={36}
          tintColor={iconColor}
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
