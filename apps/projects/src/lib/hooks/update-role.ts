import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { memberKeys } from "@/constants/keys";
import type { Member, UserRole } from "@/types";
import { updateUserRoleAction } from "@/lib/actions/workspaces/update-role";

export const useUpdateRoleMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: UserRole }) =>
      updateUserRoleAction({ userId, role }),
    onMutate: async ({ userId, role }) => {
      await queryClient.cancelQueries({ queryKey: memberKeys.lists() });
      const previousMembers = queryClient.getQueryData<Member[]>(
        memberKeys.lists(),
      );
      const previousMember = previousMembers?.find(
        (member) => member.id === userId,
      );
      if (previousMember) {
        const member: Member = { ...previousMember, role };
        queryClient.setQueryData(memberKeys.lists(), [
          ...(previousMembers || []),
          member,
        ]);
      }
      return { previousMembers };
    },
    onError: (error, variables, context) => {
      queryClient.setQueryData(memberKeys.lists(), context?.previousMembers);
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
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: memberKeys.lists() });
    },
  });

  return mutation;
};
