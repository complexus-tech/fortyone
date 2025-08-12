import type { MayaUIMessage } from "@/lib/ai/tools/types";

export type MayaChatConfig = {
  currentChatId: string;
  initialMessages?: MayaUIMessage[];
  hasSelectedChat?: boolean;
  isNewChat?: boolean;
};

export type MayaNavigationConfig = {
  chatRef?: string;
};
