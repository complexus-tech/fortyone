import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { memberKeys } from "@/constants/keys";
import type { Member } from "@/types";
import { removeMemberAction } from "@/lib/actions/workspaces/remove-member";

export const useRemoveMemberMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (memberId: string) => removeMemberAction(memberId, workspaceSlug),
    onMutate: async (memberId) => {
      await queryClient.cancelQueries({ queryKey: memberKeys.lists(workspaceSlug) });
      await queryClient.cancelQueries({ queryKey: memberKeys.all(workspaceSlug) });
      const previousMembers = queryClient.getQueryData<Member[]>(
        memberKeys.lists(workspaceSlug),
      );
      queryClient.setQueryData(memberKeys.lists(workspaceSlug), (old: Member[]) =>
        old.filter((member) => member.id !== memberId),
      );

      return { previousMembers };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(memberKeys.lists(workspaceSlug), context?.previousMembers);
      toast.error("Error", {
        description: error.message || "Failed to remove member",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({ queryKey: memberKeys.lists(workspaceSlug) });
      queryClient.invalidateQueries({ queryKey: memberKeys.all(workspaceSlug) });
    },
  });

  return mutation;
};
