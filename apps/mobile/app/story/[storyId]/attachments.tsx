import React from "react";
import { SafeContainer, Text, Row } from "@/components/ui";

export default function Attachments() {
  return (
    <SafeContainer>
      <Row className="pb-2" asContainer justify="between" align="center">
        <Text fontSize="2xl" fontWeight="semibold">
          Attachments
        </Text>
      </Row>

      <Text color="muted" className="mt-4">
        This is the Attachments tab for the story details page.
      </Text>
    </SafeContainer>
  );
}
