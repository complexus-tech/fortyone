import React from "react";
import { SafeContainer, Text } from "@/components/ui";
import { Header } from "@/modules/story/components/header";

export default function Attachments() {
  return (
    <SafeContainer>
      <Header />
      <Text color="muted" className="mt-4">
        This is the Attachments tab for the story details page.
      </Text>
    </SafeContainer>
  );
}
