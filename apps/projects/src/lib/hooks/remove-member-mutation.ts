import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { memberKeys } from "@/constants/keys";
import type { Member } from "@/types";
import { removeMemberAction } from "@/lib/actions/workspaces/remove-member";

export const useRemoveMemberMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (memberId: string) => removeMemberAction(memberId),
    onMutate: async (memberId) => {
      await queryClient.cancelQueries({ queryKey: memberKeys.lists() });
      await queryClient.cancelQueries({ queryKey: memberKeys.all });
      const previousMembers = queryClient.getQueryData<Member[]>(
        memberKeys.lists(),
      );
      queryClient.setQueryData(memberKeys.lists(), (old: Member[]) =>
        old.filter((member) => member.id !== memberId),
      );

      return { previousMembers };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(memberKeys.lists(), context?.previousMembers);
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
      queryClient.invalidateQueries({ queryKey: memberKeys.lists() });
      queryClient.invalidateQueries({ queryKey: memberKeys.all });
    },
  });

  return mutation;
};
