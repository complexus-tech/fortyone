import type { MayaUIMessage } from "@/lib/ai/tools/types";

export type MayaChatConfig = {
  currentChatId: string;
  initialMessages?: MayaUIMessage[];
  hasSelectedChat?: boolean;
  isNewChat?: boolean;
  updateChatRef: (chatId: string) => void;
  clearChatRef: (nextChatId?: string) => void;
};

export type MayaNavigationConfig = {
  chatRef?: string;
};
