import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useWorkspacePath } from "@/hooks";
import { aiChatKeys } from "../constants";
import { createAiChatAction } from "../actions/create-ai-chat";
import type { CreateAiChatPayload } from "../types";

export const useCreateAiChat = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (payload: CreateAiChatPayload) =>
      createAiChatAction(payload, workspaceSlug),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: aiChatKeys.lists() });
    },
  });
};
