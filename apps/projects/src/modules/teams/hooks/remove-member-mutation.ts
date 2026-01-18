import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useSession } from "next-auth/react";
import { useParams, useRouter } from "next/navigation";
import { useWorkspacePath } from "@/hooks";
import { memberKeys, statusKeys, teamKeys } from "@/constants/keys";
import type { Member } from "@/types";
import { storyKeys } from "@/modules/stories/constants";
import { removeTeamMemberAction } from "../actions/remove-team-member";
import { useAddMemberMutation } from "./add-member-mutation";

export const useRemoveMemberMutation = () => {
  const queryClient = useQueryClient();
  const { teamId: teamIdParam } = useParams<{ teamId?: string }>();
  const { data: session } = useSession();
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const router = useRouter();
  const currentUserId = session?.user?.id ?? "";
  const toastId = "remove-member-mutation";
  const { mutate: addMember } = useAddMemberMutation();

  const mutation = useMutation({
    mutationFn: ({ teamId, memberId }: { teamId: string; memberId: string }) =>
      removeTeamMemberAction(teamId, memberId, workspaceSlug),
    onMutate: async ({ teamId, memberId }) => {
      toast.loading(
        memberId === currentUserId ? "Leaving team..." : "Removing member...",
        {
          id: toastId,
          description: "Please wait...",
        },
      );
      await queryClient.cancelQueries({ queryKey: memberKeys.team(workspaceSlug, teamId) });
      const previousMembers = queryClient.getQueryData<Member[]>(
        memberKeys.team(workspaceSlug, teamId),
      );
      if (previousMembers) {
        queryClient.setQueryData(memberKeys.team(workspaceSlug, teamId), (old: Member[]) =>
          old.filter((member) => member.id !== memberId),
        );
      }

      return { previousMembers };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(
        memberKeys.team(workspaceSlug, variables.teamId),
        context?.previousMembers,
      );
      toast.error(
        variables.memberId === currentUserId
          ? "Failed to leave team"
          : "Failed to remove member",
        {
          description:
            error.message ||
            (variables.memberId === currentUserId
              ? "Failed to leave team"
              : "Failed to remove member"),
          id: toastId,
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
        memberId === currentUserId ? "Left team" : "Removed member",
        {
          id: toastId,
          description:
            memberId === currentUserId
              ? "You are no longer a member of this team."
              : "Member removed",
          action: {
            label: memberId === currentUserId ? "Rejoin" : "Undo",
            onClick: () => {
              addMember({ teamId, memberId });
            },
          },
        },
      );
      queryClient.invalidateQueries({ queryKey: memberKeys.team(workspaceSlug, teamId) });
      if (memberId === session?.user?.id) {
        queryClient.invalidateQueries({ queryKey: teamKeys.lists(workspaceSlug) });
        queryClient.invalidateQueries({ queryKey: teamKeys.public(workspaceSlug) });
        queryClient.invalidateQueries({ queryKey: storyKeys.mine(workspaceSlug) });
        queryClient.invalidateQueries({ queryKey: statusKeys.lists(workspaceSlug) });
        if (teamIdParam === teamId) {
          router.push(withWorkspace("/my-work"));
        }
      }
    },
  });

  return mutation;
};
