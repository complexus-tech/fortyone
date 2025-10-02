import React from "react";
import { SafeContainer, Text, Row, Back } from "@/components/ui";

export default function Attachments() {
  return (
    <SafeContainer>
      <Row className="pb-2" gap={2} align="center">
        <Back />
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
