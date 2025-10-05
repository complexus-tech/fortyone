import React from "react";
import { Attachments } from "@/modules/story/attachments";
import { SafeContainer } from "@/components/ui";
import { Header } from "@/modules/story/components/header";

export default function AttachmentsPage() {
  return (
    <SafeContainer isFull>
      <Header />
      <Attachments />
    </SafeContainer>
  );
}
