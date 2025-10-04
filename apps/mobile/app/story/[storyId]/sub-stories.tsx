import React from "react";
import { SubStories } from "@/modules/story/sub-stories";
import { SafeContainer } from "@/components/ui";
import { Header } from "@/modules/story/components/header";

export default function SubStoriesPage() {
  return (
    <SafeContainer isFull>
      <Header />
      <SubStories />
    </SafeContainer>
  );
}
