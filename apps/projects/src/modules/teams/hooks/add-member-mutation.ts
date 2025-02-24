import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { memberKeys } from "@/constants/keys";
import type { Member } from "@/types";
import { useMembers } from "@/lib/hooks/members";
import { addTeamMemberAction } from "../actions/add-team-member";

export const useAddMemberMutation = () => {
  const queryClient = useQueryClient();
  const { data: members = [] } = useMembers();

  const mutation = useMutation({
    mutationFn: ({ teamId, memberId }: { teamId: string; memberId: string }) =>
      addTeamMemberAction(teamId, memberId),
    onMutate: async ({ teamId, memberId }) => {
      await queryClient.cancelQueries({ queryKey: memberKeys.team(teamId) });
      const previousMembers = queryClient.getQueryData<Member[]>(
        memberKeys.team(teamId),
      );
      const newMember = members.find((member) => member.id === memberId);
      if (newMember) {
        const newMembers = [...(previousMembers || []), newMember].sort(
          (a, b) => a.fullName.localeCompare(b.fullName),
        );
        queryClient.setQueryData(memberKeys.team(teamId), newMembers);
      }
      return { previousMembers };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(
        memberKeys.team(variables.teamId),
        context?.previousMembers,
      );
      toast.error("Error", {
        description: error.message || "Failed to add member",
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
