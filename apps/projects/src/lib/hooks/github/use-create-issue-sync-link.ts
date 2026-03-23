import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { githubKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { createGitHubIssueSyncLinkAction } from "@/lib/actions/github/create-issue-sync-link";
import type {
  CreateGitHubIssueSyncLinkInput,
  GitHubIntegration,
} from "@/modules/settings/workspace/integrations/github/types";

export const useCreateGitHubIssueSyncLink = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = githubKeys.integration(workspaceSlug);

  return useMutation({
    mutationFn: (input: CreateGitHubIssueSyncLinkInput) =>
      createGitHubIssueSyncLinkAction(input, workspaceSlug),
    onMutate: async (input) => {
      await queryClient.cancelQueries({ queryKey });
      const previous = queryClient.getQueryData<GitHubIntegration>(queryKey);

      queryClient.setQueryData<GitHubIntegration>(queryKey, (old) => {
        if (!old) return old;

        const repository = old.repositories.find(
          (item) => item.id === input.repositoryId,
        );
        const existingTeamLink = old.issueSyncLinks.find(
          (item) => item.teamId === input.teamId,
        );

        return {
          ...old,
          issueSyncLinks: [
            ...old.issueSyncLinks,
            {
              id: `draft-${input.repositoryId}-${input.teamId}`,
              repositoryId: input.repositoryId,
              repositoryName: repository?.fullName ?? "Repository",
              teamId: input.teamId,
              teamName: existingTeamLink?.teamName ?? "Team",
              teamColor: existingTeamLink?.teamColor ?? "#6B7280",
              syncDirection: input.syncDirection,
              isActive: true,
            },
          ],
        };
      });

      return { previous };
    },
    onError: (_err, _input, context) => {
      if (context?.previous) {
        queryClient.setQueryData(queryKey, context.previous);
      }
      toast.error("GitHub", { description: "Failed to link repository" });
    },
    onSuccess: (res, _input, context) => {
      if (res.error?.message) {
        if (context?.previous) {
          queryClient.setQueryData(queryKey, context.previous);
        }
        toast.error("GitHub", { description: res.error.message });
        return;
      }
      toast.success("Repository linked");
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });
};
