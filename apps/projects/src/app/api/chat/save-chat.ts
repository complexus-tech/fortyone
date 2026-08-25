import type { UIMessage } from "ai";
import { saveAiChatMessagesAction } from "@/modules/ai-chats/actions/save-ai-chat-messages";
import { createAiChatAction } from "@/modules/ai-chats/actions/create-ai-chat";
import { getChatTitle } from "./chat-title";

export const saveChat = async ({
  id,
  messages,
  workspaceSlug,
}: {
  id: string;
  messages: UIMessage[];
  workspaceSlug: string;
}) => {
  const title = messages.length <= 3 ? getChatTitle(messages) : "";

  try {
    if (title) {
      await createAiChatAction({ id, title, messages }, workspaceSlug);
    } else {
      await saveAiChatMessagesAction({ id, messages }, workspaceSlug);
    }
  } catch (error) {
    // log to analytics later
  }
};
