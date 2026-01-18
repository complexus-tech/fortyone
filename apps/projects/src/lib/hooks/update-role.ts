import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { memberKeys } from "@/constants/keys";
import type { Member, UserRole } from "@/types";
import { updateUserRoleAction } from "@/lib/actions/workspaces/update-role";

export const useUpdateRoleMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: UserRole }) =>
      updateUserRoleAction({ userId, role, workspaceSlug }),
    onMutate: async ({ userId, role }) => {
      await queryClient.cancelQueries({ queryKey: memberKeys.lists(workspaceSlug) });
      const previousMembers = queryClient.getQueryData<Member[]>(
        memberKeys.lists(workspaceSlug),
      );

      if (previousMembers) {
        queryClient.setQueryData(
          memberKeys.lists(workspaceSlug),
          previousMembers.map((member) => {
            if (member.id === userId) {
              return { ...member, role };
            }
            return member;
          }),
        );
      }
      return { previousMembers };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(memberKeys.lists(workspaceSlug), context?.previousMembers);
      toast.error("Failed to update user role", {
        description: error.message || "Failed to update user role",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },
    onSuccess: (res) => {
      if (res?.error?.message) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({ queryKey: memberKeys.lists(workspaceSlug) });
    },
  });

  return mutation;
};
