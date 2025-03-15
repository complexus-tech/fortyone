import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useSession } from "next-auth/react";
import { memberKeys, teamKeys } from "@/constants/keys";
import type { Member } from "@/types";
import { useMembers } from "@/lib/hooks/members";
import { addTeamMemberAction } from "../actions/add-team-member";

export const useAddMemberMutation = () => {
  const queryClient = useQueryClient();
  const { data: members = [] } = useMembers();
  const { data: session } = useSession();
  const currentUserId = session?.user?.id ?? "";
  const toastId = "add-member-mutation";

  const mutation = useMutation({
    mutationFn: ({ teamId, memberId }: { teamId: string; memberId: string }) =>
      addTeamMemberAction(teamId, memberId),
    onMutate: async ({ teamId, memberId }) => {
      toast.loading(
        memberId === currentUserId ? "Joining team..." : "Adding member...",
        {
          id: toastId,
          description: "Please wait...",
        },
      );
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
      toast.error(
        variables.memberId === currentUserId
          ? "Failed to join team"
          : "Failed to add member",
        {
          description:
            error.message ||
            (variables.memberId === currentUserId
              ? "Failed to join team"
              : "Failed to add member"),
          action: {
            label: "Retry",
            onClick: () => {
              mutation.mutate(variables);
            },
          },
        },
      );
    },
    onSuccess: (res, { teamId, memberId }) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      toast.success(
        memberId === currentUserId ? "Joined team" : "Added member",
        {
          id: toastId,
          description:
            memberId === currentUserId
              ? "Congratulations! You are now a member of the team."
              : "Member added",
        },
      );
      queryClient.invalidateQueries({ queryKey: memberKeys.team(teamId) });
      if (memberId === session?.user?.id) {
        queryClient.invalidateQueries({ queryKey: teamKeys.lists() });
        queryClient.invalidateQueries({ queryKey: teamKeys.public() });
      }
    },
  });

  return mutation;
};
