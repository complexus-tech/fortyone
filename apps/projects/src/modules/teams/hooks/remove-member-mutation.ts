import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { memberKeys } from "@/constants/keys";
import type { Member } from "@/types";
import { removeTeamMemberAction } from "../actions/remove-team-member";

export const useRemoveMemberMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ teamId, memberId }: { teamId: string; memberId: string }) =>
      removeTeamMemberAction(teamId, memberId),
    onMutate: async ({ teamId, memberId }) => {
      await queryClient.cancelQueries({ queryKey: memberKeys.team(teamId) });
      const previousMembers = queryClient.getQueryData<Member[]>(
        memberKeys.team(teamId),
      );
      queryClient.setQueryData(memberKeys.team(teamId), (old: Member[]) =>
        old.filter((member) => member.id !== memberId),
      );

      return { previousMembers };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(
        memberKeys.team(variables.teamId),
        context?.previousMembers,
      );
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
    onSuccess: (_, { teamId }) => {
      queryClient.invalidateQueries({ queryKey: memberKeys.team(teamId) });
    },
  });

  return mutation;
};
