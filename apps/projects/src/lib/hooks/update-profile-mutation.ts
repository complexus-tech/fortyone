import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { userKeys } from "@/constants/keys";
import type { User } from "@/types";
import { updateProfile } from "../actions/users/update";
import type { UpdateProfile } from "../actions/users/update";

export const useUpdateProfileMutation = () => {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (updates: UpdateProfile) => updateProfile(updates),

    onMutate: async (updates) => {
      await queryClient.cancelQueries({ queryKey: userKeys.profile() });
      const previousProfile = queryClient.getQueryData<User>(
        userKeys.profile(),
      );

      if (previousProfile) {
        queryClient.setQueryData<User>(userKeys.profile(), {
          ...previousProfile,
          ...updates,
        });
      }

      return { previousProfile };
    },

    onError: (_, variables, context) => {
      if (context?.previousProfile) {
        queryClient.setQueryData(userKeys.profile(), context.previousProfile);
      }
      toast.error("Failed to update profile", {
        description: "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate(variables);
          },
        },
      });
    },

    onSuccess: () => {
      toast.success("Success", {
        description: "Your profile has been updated",
      });
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: userKeys.profile() });
    },
  });

  return mutation;
};
