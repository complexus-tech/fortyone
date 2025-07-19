import type { Message } from "@ai-sdk/react";

export type MayaChatConfig = {
  currentChatId: string;
  initialMessages?: Message[];
  hasSelectedChat?: boolean;
  isNewChat?: boolean;
};

export type MayaNavigationConfig = {
  chatRef?: string;
};
