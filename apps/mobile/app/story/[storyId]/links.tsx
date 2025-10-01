import React from "react";
import { SafeContainer, Text, Row, Back } from "@/components/ui";

export default function Links() {
  return (
    <SafeContainer>
      <Row className="pb-2" gap={2} align="center">
        <Back />
        <Text fontSize="2xl" fontWeight="semibold">
          Links
        </Text>
      </Row>

      <Text color="muted" className="mt-4">
        This is the Links tab for the story details page.
      </Text>
    </SafeContainer>
  );
}
