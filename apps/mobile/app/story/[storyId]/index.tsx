import React from "react";
import { SafeContainer } from "@/components/ui";
import { Header } from "@/modules/story/components/header";
import { Story } from "@/modules/story";

export default function StoryOverview() {
  return (
    <SafeContainer isFull>
      <Header />
      <Story />
    </SafeContainer>
  );
}
