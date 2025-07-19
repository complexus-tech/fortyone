"use client";

import { MayaChat, useMayaNavigation } from "@/modules/maya";

export default function MayaPage() {
  const { chatRef, getInitialChatId, isNewChat } = useMayaNavigation();

  const isNewChatState = isNewChat();

  return (
    <MayaChat
      config={{
        currentChatId: getInitialChatId(),
        hasSelectedChat: Boolean(chatRef),
        isNewChat: isNewChatState,
      }}
    />
  );
}
