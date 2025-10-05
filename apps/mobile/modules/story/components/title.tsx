import React from "react";
import { Text, Row } from "@/components/ui";

export const Title = ({ title }: { title?: string }) => {
  return (
    <Row asContainer>
      <Text fontSize="xl" fontWeight="semibold">
        {title}
      </Text>
    </Row>
  );
};
