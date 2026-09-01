import type { MayaUIMessage } from "@/lib/ai/tools/types";

export type MayaChatConfig = {
  currentChatId: string;
  initialMessages?: MayaUIMessage[];
  hasSelectedChat: boolean;
  onUserMessageSubmitted?: () => void;
  updateChatRef: (chatId: string) => void;
  clearChatRef: (nextChatId?: string) => void;
};

export type MayaNavigationConfig = {
  chatRef?: string;
};
