import { useMutation, useQueryClient } from "@tanstack/react-query";
import { aiChatKeys } from "../constants";
import { deleteAiChatAction } from "../actions/delete-ai-chat";

export const useDeleteAiChat = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteAiChatAction(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: aiChatKeys.lists() });
      queryClient.removeQueries({ queryKey: aiChatKeys.detail(id) });
      queryClient.removeQueries({ queryKey: aiChatKeys.messages(id) });
    },
  });
};
