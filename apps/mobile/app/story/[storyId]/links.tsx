import React from "react";
import { SafeContainer, Text } from "@/components/ui";
import { Header } from "@/modules/story/components/header";

export default function Links() {
  return (
    <SafeContainer>
      <Header />

      <Text color="muted" className="mt-4">
        This is the Links tab for the story details page.
      </Text>
    </SafeContainer>
  );
}
