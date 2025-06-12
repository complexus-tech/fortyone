import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { deleteProfileImageAction } from "@/lib/actions/users/delete-profile-image";
import { userKeys } from "@/constants/keys";
import type { User } from "@/types";

export const useDeleteProfileImageMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: () => deleteProfileImageAction(),
    onMutate: async () => {
      await queryClient.cancelQueries({ queryKey: userKeys.profile() });
      const previousProfile = queryClient.getQueryData<User>(
        userKeys.profile(),
      );

      if (previousProfile) {
        queryClient.setQueryData<User>(userKeys.profile(), {
          ...previousProfile,
          avatarUrl: null,
        });
        return { previousProfile };
      }
    },
    onError: (error, _, context) => {
      if (context?.previousProfile) {
        queryClient.setQueryData(userKeys.profile(), context.previousProfile);
      }
      toast.error("Failed to remove profile image", {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate();
          },
        },
      });
    },
    onSuccess: (res) => {
      if (res.error?.message) {
        throw new Error(res.error.message);
      }
      queryClient.invalidateQueries({ queryKey: userKeys.profile() });
    },
  });

  return mutation;
};
