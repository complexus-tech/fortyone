import React from "react";
import { Links } from "@/modules/story/links";
import { SafeContainer } from "@/components/ui";
import { Header } from "@/modules/story/components/header";

export default function LinksPage() {
  return (
    <SafeContainer isFull>
      <Header />
      <Links />
    </SafeContainer>
  );
}
