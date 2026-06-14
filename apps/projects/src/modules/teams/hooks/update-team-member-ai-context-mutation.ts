import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { memberKeys } from "@/constants/keys";
import type { Member } from "@/types";
import { updateTeamMemberAIContextAction } from "../actions/update-team-member-ai-context";

type UpdateTeamMemberAIContextInput = {
  teamId: string;
  memberId: string;
  roleTitle: string;
  roleDescription: string;
};

export const useUpdateTeamMemberAIContextMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: ({
      teamId,
      memberId,
      roleTitle,
      roleDescription,
    }: UpdateTeamMemberAIContextInput) =>
      updateTeamMemberAIContextAction(
        teamId,
        memberId,
        { roleTitle, roleDescription },
        workspaceSlug,
      ),
    onMutate: async ({ teamId, memberId, roleTitle, roleDescription }) => {
      await queryClient.cancelQueries({
        queryKey: memberKeys.team(workspaceSlug, teamId),
      });

      const previousMembers = queryClient.getQueryData<Member[]>(
        memberKeys.team(workspaceSlug, teamId),
      );

      queryClient.setQueryData<Member[]>(
        memberKeys.team(workspaceSlug, teamId),
        (members = []) =>
          members.map((member) =>
            member.id === memberId
              ? {
                  ...member,
                  teamAiRoleTitle: roleTitle,
                  teamAiRoleDescription: roleDescription,
                }
              : member,
          ),
      );

      return { previousMembers };
    },
    onError: (error, variables, context) => {
      if (context?.previousMembers) {
        queryClient.setQueryData(
          memberKeys.team(workspaceSlug, variables.teamId),
          context.previousMembers,
        );
      }

      toast.error("Failed to update member role context", {
        description: error.message || "Your changes were not saved",
      });
    },
    onSuccess: (res, variables) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }

      queryClient.invalidateQueries({
        queryKey: memberKeys.team(workspaceSlug, variables.teamId),
      });
      toast.success("Member role context updated");
    },
  });
};
