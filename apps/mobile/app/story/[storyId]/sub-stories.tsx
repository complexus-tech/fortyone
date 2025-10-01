import React from "react";
import { SafeContainer, Text, Row, Back } from "@/components/ui";

export default function SubStories() {
  return (
    <SafeContainer>
      <Row className="pb-2" gap={2} align="center">
        <Back />
        <Text fontSize="2xl" fontWeight="semibold">
          Sub Stories
        </Text>
      </Row>

      <Text color="muted" className="mt-4">
        This is the Sub Stories tab for the story details page.
      </Text>
    </SafeContainer>
  );
}
