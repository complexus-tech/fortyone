import { useMutation, useQueryClient } from "@tanstack/react-query";
import { aiChatKeys } from "../constants";
import { saveAiChatMessagesAction } from "../actions/save-ai-chat-messages";
import type { SaveMessagesPayload } from "../types";

export const useSaveAiChatMessages = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: SaveMessagesPayload) =>
      saveAiChatMessagesAction(payload),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: aiChatKeys.messages(id) });
      queryClient.invalidateQueries({ queryKey: aiChatKeys.detail(id) });
    },
  });
};
