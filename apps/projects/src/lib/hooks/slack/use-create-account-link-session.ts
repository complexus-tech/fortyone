import { useMutation } from "@tanstack/react-query";
import { useWorkspacePath } from "@/hooks";
import { createSlackAccountLinkSessionAction } from "@/lib/actions/slack/create-account-link-session";

export const useCreateSlackAccountLinkSession = () => {
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (returnUrl: string) =>
      createSlackAccountLinkSessionAction(workspaceSlug, returnUrl),
  });
};
