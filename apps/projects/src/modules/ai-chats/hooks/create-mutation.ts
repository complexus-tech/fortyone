import { useMutation, useQueryClient } from "@tanstack/react-query";
import { aiChatKeys } from "../constants";
import { createAiChatAction } from "../actions/create-ai-chat";
import type { CreateAiChatPayload } from "../types";

export const useCreateAiChat = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: CreateAiChatPayload) => createAiChatAction(payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: aiChatKeys.lists() });
    },
  });
};
