import { useMutation, useQueryClient } from "@tanstack/react-query";
import { aiChatKeys } from "../constants";
import { updateAiChatAction } from "../actions/update-ai-chat";
import type { UpdateAiChatPayload } from "../types";

export const useUpdateAiChat = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      id,
      payload,
    }: {
      id: string;
      payload: UpdateAiChatPayload;
    }) => updateAiChatAction(id, payload),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: aiChatKeys.lists() });
      queryClient.invalidateQueries({ queryKey: aiChatKeys.detail(id) });
    },
  });
};
