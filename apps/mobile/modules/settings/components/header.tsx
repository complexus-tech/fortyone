import React from "react";
import { Row, Text } from "@/components/ui";

export const Header = () => {
  return (
    <Row className="mb-3" asContainer justify="between" align="center">
      <Text fontSize="3xl" fontWeight="semibold">
        Settings
      </Text>
    </Row>
  );
};
