import React from "react";
import { SafeContainer, Text } from "@/components/ui";
import { Header } from "@/modules/story/components/header";

export default function SubStories() {
  return (
    <SafeContainer>
      <Header />

      <Text color="muted" className="mt-4">
        This is the Sub Stories tab for the story details page.
      </Text>
    </SafeContainer>
  );
}
