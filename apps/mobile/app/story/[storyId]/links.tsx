import React from "react";
import { SafeContainer, Text, Row } from "@/components/ui";

export default function Links() {
  return (
    <SafeContainer edges={[]}>
      <Row className="pb-2" justify="between" align="center">
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
